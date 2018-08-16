# mxnet-job

> Prototypes for running MXNet jobs.


* [Installing MXNet Operator](#install-mxnet-operator)
* [Verify that MXNet support is included in your Kubeflow deployment](#verify-deployment)
* [Creating a MXNet Job](#create-mxnet-job)
* [Monitoring a MXNet Job](#monitor-mxnet-job)

## Installing MXNet Operator

If you haven’t already done so please follow [the Getting Started Guide](https://www.kubeflow.org/docs/started/getting-started/) to deploy Kubeflow.

An alpha version of MXNet support was introduced with Kubeflow 0.2.0. You must be using a version of Kubeflow newer than 0.2.0.

## Verify Deployment

Check that the PyTorch custom resource is installed

```
kubectl get crd
```

The output should include mxjobs.kubeflow.org

```
NAME                                           AGE
...
mxjobs.kubeflow.org                       4d
...
```

If it is not included you can add it as follows

```
cd ${KSONNET_APP}
ks pkg install kubeflow/mxnet-job
ks generate mxnet-operator mxnet-operator
ks apply ${ENVIRONMENT} -c mxnet-operator
```


## Create MXNet Job

You can create MXNet Job by defining a MXNetjob config file. See [distributed MNIST](https://github.com/TuSimple/mxnet-operator/blob/master/examples/mxjob_sample/gpu/mx_job_dist.yaml) example config file. You may change the config file based on your requirements.

```
cat examples/mxjob_sample/gpu/mx_job_dist.yaml
```

Deploy the MXNetJob resource to start training:

```
kubectl create -f examples/mxjob_sample/gpu/mx_job_dist.yaml
```

You should now be able to see the created pods matching the specified number of replicas.


## Monitor MXNet Job

```
kubectl get -o yaml mxjobs ${JOB_NAME}
```

See the status section to monitor the job status. Here is sample output when the job is successfully running.

```
apiVersion: kubeflow.org/v1alpha1
kind: MXJob
metadata:
  clusterName: ""
  creationTimestamp: 2018-08-16T08:05:50Z
  generation: 1
  name: gpu-dist-job
  namespace: default
  resourceVersion: "109005"
  selfLink: /apis/kubeflow.org/v1alpha1/namespaces/default/mxjobs/gpu-dist-job
  uid: 30a876eb-a12b-11e8-8432-704d7b2c0a63
spec:
  RuntimeId: i4i9
  jobMode: dist
  mxImage: jzp1025/mxnet:test
  replicaSpecs:
  - PsRootPort: 9000
    mxReplicaType: SCHEDULER
    replicas: 1
    template:
      metadata:
        creationTimestamp: null
      spec:
        containers:
        - args:
          - train_mnist.py
          command:
          - python
          image: jzp1025/mxnet:gpu_job
          name: mxnet
          resources:
            limits:
              nvidia.com/gpu: "1"
          workingDir: /incubator-mxnet/example/image-classification
        restartPolicy: OnFailure
  - PsRootPort: 9091
    mxReplicaType: SERVER
    replicas: 2
    template:
      metadata:
        creationTimestamp: null
      spec:
        containers:
        - args:
          - train_mnist.py
          command:
          - python
          image: jzp1025/mxnet:gpu_job
          name: mxnet
          resources:
            limits:
              nvidia.com/gpu: "1"
          workingDir: /incubator-mxnet/example/image-classification
        restartPolicy: OnFailure
  - PsRootPort: 9091
    mxReplicaType: WORKER
    replicas: 2
    template:
      metadata:
        creationTimestamp: null
      spec:
        containers:
        - args:
          - train_mnist.py
          - --num-epochs
          - "10"
          - --num-layers
          - "2"
          - --kv-store
          - dist_device_sync
          - --gpus
          - "0"
          command:
          - python
          image: jzp1025/mxnet:gpu_job
          name: mxnet
          resources:
            limits:
              nvidia.com/gpu: "1"
          workingDir: /incubator-mxnet/example/image-classification
        restartPolicy: OnFailure
  terminationPolicy:
    chief:
      replicaIndex: 0
      replicaName: SCHEDULER
status:
  phase: Running
  reason: ""
  replicaStatuses:
  - ReplicasStates:
      Running: 1
    mx_replica_type: SCHEDULER
    state: Running
  - ReplicasStates:
      Running: 2
    mx_replica_type: SERVER
    state: Running
  - ReplicasStates:
      Running: 2
    mx_replica_type: WORKER
    state: Running
  state: Running
```


