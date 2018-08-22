package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	mxv1alpha1 "github.com/kubeflow/mxnet-operator/pkg/apis/mxnet/v1alpha1"
	mxjobclient "github.com/kubeflow/mxnet-operator/pkg/client/clientset/versioned"
	"github.com/kubeflow/mxnet-operator/pkg/util"
	"k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	name      = flag.String("name", "", "The name for the MXJob to create..")
	namespace = flag.String("namespace", "default", "The namespace to create the test job in.")
	numJobs   = flag.Int("num_jobs", 1, "The number of jobs to run.")
	timeout   = flag.Duration("timeout", 10*time.Minute, "The timeout for the test")
	image     = flag.String("image", "", "The Test image to run")
)

type mxReplicaType mxv1alpha1.MXReplicaType

func (mxrt mxReplicaType) toSpec(replica int32) *mxv1alpha1.MXReplicaSpec {
	return &mxv1alpha1.MXReplicaSpec{
		Replicas:      proto.Int32(replica),
		PsRootPort:    proto.Int32(9001),
		MXReplicaType: mxv1alpha1.MXReplicaType(mxrt),
		Template: &v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "mxnet",
						Image:           *image,
						ImagePullPolicy: "IfNotPresent",
					},
				},
				RestartPolicy: v1.RestartPolicyOnFailure,
			},
		},
	}

}

func run() (string, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	if *name == "" {
		name = proto.String("example-job")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	if *image == "" {
		log.Fatalf("--image must be provided.")
	}

	// create the clientset
	client := kubernetes.NewForConfigOrDie(config)

	mxJobClient, err := mxjobclient.NewForConfig(config)
	if err != nil {
		return "", err
	}

	original := &mxv1alpha1.MXJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: *name,
			Labels: map[string]string{
				"test.mlkube.io": "",
			},
		},
		Spec: mxv1alpha1.MXJobSpec{
			ReplicaSpecs: []*mxv1alpha1.MXReplicaSpec{
				mxReplicaType(mxv1alpha1.SCHEDULER).toSpec(1),
				mxReplicaType(mxv1alpha1.SERVER).toSpec(2),
				mxReplicaType(mxv1alpha1.WORKER).toSpec(2),
			},
		},
	}

	// Create MXJob
	_, err = mxJobClient.KubeflowV1alpha1().MXJobs(*namespace).Create(original)
	if err != nil {
		log.Errorf("Creating the job failed; %v", err)
		return *name, err
	}

	// TODO(jose5918) Wait for completed state
	// Wait for operator to reach running state
	var mxJob *mxv1alpha1.MXJob
	for endTime := time.Now().Add(*timeout); time.Now().Before(endTime); {
		mxJob, err = mxJobClient.KubeflowV1alpha1().MXJobs(*namespace).Get(*name, metav1.GetOptions{})
		if err != nil {
			log.Warningf("There was a problem getting MXJob: %v; error %v", *name, err)
		}

		if mxJob.Status.State == mxv1alpha1.StateSucceeded || mxJob.Status.State == mxv1alpha1.StateFailed {
			log.Infof("job %v finished:\n%v", *name, util.Pformat(mxJob))
			break
		}
		log.Infof("Waiting for job %v to finish:\n%v", *name, util.Pformat(mxJob))
		time.Sleep(5 * time.Second)
	}

	if mxJob == nil {
		return *name, fmt.Errorf("Failed to get MXJob %v", *name)
	}

	if mxJob.Status.State != mxv1alpha1.StateSucceeded {
		// TODO(jlewi): Should we clean up the job.
		return *name, fmt.Errorf("MXJob %v did not succeed;\n %v", *name, util.Pformat(mxJob))
	}

	if mxJob.Spec.RuntimeId == "" {
		return *name, fmt.Errorf("MXJob %v doesn't have a RuntimeId", *name)
	}

	// Loop over each replica and make sure the expected resources were created.
	for _, r := range original.Spec.ReplicaSpecs {
		baseName := strings.ToLower(string(r.MXReplicaType))

		for i := 0; i < int(*r.Replicas); i++ {
			jobName := fmt.Sprintf("%v-%v-%v-%v", fmt.Sprintf("%.40s", original.ObjectMeta.Name), baseName, mxJob.Spec.RuntimeId, i)

			_, err := mxJobClient.KubeflowV1alpha1().MXJobs(*namespace).Get(*name, metav1.GetOptions{})

			if err != nil {
				return *name, fmt.Errorf("MXJob %v did not create Job %v for ReplicaType %v Index %v", *name, jobName, r.MXReplicaType, i)
			}
		}
	}

	// Delete the job and make sure all subresources are properly garbage collected.
	if err := mxJobClient.KubeflowV1alpha1().MXJobs(*namespace).Delete(*name, &metav1.DeleteOptions{}); err != nil {
		log.Fatalf("Failed to delete MXJob %v; error %v", *name, err)
	}

	// Define sets to keep track of Job controllers corresponding to Replicas
	// that still exist.
	jobs := make(map[string]bool)

	// Loop over each replica and make sure the expected resources are being deleted.
	for _, r := range original.Spec.ReplicaSpecs {
		baseName := strings.ToLower(string(r.MXReplicaType))

		for i := 0; i < int(*r.Replicas); i++ {
			jobName := fmt.Sprintf("%v-%v-%v-%v", fmt.Sprintf("%.40s", original.ObjectMeta.Name), baseName, mxJob.Spec.RuntimeId, i)

			jobs[jobName] = true
		}
	}

	// Wait for all jobs and deployment to be deleted.
	for endTime := time.Now().Add(*timeout); time.Now().Before(endTime) && len(jobs) > 0; {
		for k := range jobs {
			_, err := client.BatchV1().Jobs(*namespace).Get(k, metav1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				// Deleting map entry during loop is safe.
				// See: https://stackoverflow.com/questions/23229975/is-it-safe-to-remove-selected-keys-from-golang-map-within-a-range-loop
				delete(jobs, k)
			} else {
				log.Infof("Job %v still exists", k)
			}
		}

		if len(jobs) > 0 {
			time.Sleep(5 * time.Second)
		}
	}

	if len(jobs) > 0 {
		return *name, fmt.Errorf("Not all Job controllers were successfully deleted for MXJob %v.", *name)
	}

	return *name, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func runCmd(cmd *exec.Cmd) error {
	var waitStatus syscall.WaitStatus
	err := cmd.Run()
	if err != nil {
		// Did the command fail because of an unsuccessful exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			output, _ := cmd.CombinedOutput()
			log.Infof("exitcode %d: %s", waitStatus.ExitStatus(), string(output))
		}
	} else {
		// Command was successful
		_ = cmd.ProcessState.Sys().(syscall.WaitStatus)
	}
	return err
}

func main() {
	flag.Parse()

	type Result struct {
		Error error
		Name  string
	}
	c := make(chan Result)

	for i := 0; i < *numJobs; i++ {
		go func() {
			name, err := run()
			if err != nil {
				log.Errorf("Job %v didn't run successfully; %v", name, err)
			} else {
				log.Infof("Job %v ran successfully", name)
			}
			c <- Result{
				Name:  name,
				Error: err,
			}
		}()
	}

	numSucceded := 0
	numFailed := 0

	for endTime := time.Now().Add(*timeout); numSucceded+numFailed < *numJobs && time.Now().Before(endTime); {
		select {
		case res := <-c:
			if res.Error == nil {
				numSucceded += 1
			} else {
				numFailed += 1
			}
		case <-time.After(endTime.Sub(time.Now())):
			log.Errorf("Timeout waiting for MXJob to finish.")
			fmt.Println("timeout 2")
		}
	}

	if numSucceded+numFailed < *numJobs {
		log.Errorf("Timeout waiting for jobs to finish; only %v of %v MXJobs completed.", numSucceded+numFailed, *numJobs)
	}

	// Generate TAP (https://testanything.org/) output
	fmt.Println("1..1")
	if numSucceded == *numJobs {
		fmt.Println("ok 1 - Successfully ran MXJob")
	} else {
		fmt.Printf("not ok 1 - Running MXJobs failed \n")
		// Exit with non zero exit code for Helm tests.
		os.Exit(1)
	}
}
