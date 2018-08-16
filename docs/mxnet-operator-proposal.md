<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Motivation](#motivation)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [API (CRD and MXJob)](#api-crd-and-MXJob)
  - [Custom Resource Definition](#custom-resource-definition)
  - [MXJob Example](#mxjob-example)
- [Design](#design)
- [User Guide](#user-guide)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

_Status_

* 2018-08-13 - cpu version

## Motivation
MXNet is a popular machine learning framework which currently does not have an operator/controller for Kubernetes. This proposal is aimed at defining what that operator should look like, and adding it to Kubeflow.

## Goals
A Kubeflow user should be able to run training using MXNet as easily as then can using Tensorflow.  This proposal is centered around a Kubernetes operator for MXNet. A user should be able to run both single node and distributed training jobs with MXNet.

This proposal defines the following:
- A MXNet operator
- A way to deploy the operator with ksonnet
- A distributed MXNet example

## Non-Goals
For the scope of this proposal, we won't be addressing the method for serving the model.

## API (CRD and MXJob)

### Custom Resource Definition
The custom resource submitted to the Kubernetes API would look something like this:
```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: mxjobs.kubeflow.org
spec:
  group: kubeflow.org
  # list of versions supported by this CustomResourceDefinition
  version: v1alpha1
  names:
    plural: mxjobs
    # singular name to be used as an alias on the CLI and for display
    singular: mxjob
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: MXJob

```

This MXNetJob resembles the existing TFJob for the tf-operator.  The main differences being the new Scheduler role , and the changes of the environment options.

Job Roles Contents : 1 Scheduler , N Server , M worker

### MXJob Example
```yaml
apiVersion: "kubeflow.org/v1alpha1"
kind: "MXJob"
metadata:
  name: "example-dist-job"
spec:
  jobMode: "dist"
  replicaSpecs:
    - replicas: 1
      mxReplicaType: SCHEDULER
      PsRootPort: 9000
      template:
        spec:
          containers:
            - image: jzp1025/mxnet:test
              name: mxnet
              command: ["python"]
              args: ["train_mnist.py"]
              workingDir: "/incubator-mxnet/example/image-classification"
          restartPolicy: OnFailure
    - replicas: 1 
      mxReplicaType: SERVER
      template:
        spec:
          containers:
            - image: jzp1025/mxnet:test
              name: mxnet
              command: ["python"]
              args: ["train_mnist.py"]
              workingDir: "/incubator-mxnet/example/image-classification"
          restartPolicy: OnFailure
    - replicas: 1
      mxReplicaType: WORKER
      template:
        spec:
          containers:
            - image: jzp1025/mxnet:test
              name: mxnet
              command: ["python"]
              args: ["train_mnist.py","--num-epochs=10","--num-layers=2","--kv-store=dist_device_sync"]
              workingDir: "/incubator-mxnet/example/image-classification"
          restartPolicy: OnFailure
```

The environment variables will be set in each pod due to the mxReplicaType when initializing a distributed process group with MXNet. There must be 1 and only 1 scheduler in the job.

## Design
This is an implementaion of the MXNet distributed design patterns, found [here](https://mxnet.incubator.apache.org/versions/master/faq/model_parallel_lstm.html), via the lense of TFJob found [here](https://github.com/kubeflow/tf-operator). In the case of Kubernetes, because the operator is able to easily apply configurations to each process, we will use the environment variable initialization method found [here](https://mxnet.incubator.apache.org/versions/master/faq/distributed_training.html).

## User Guide
Please refer to the User Guide, found [here](https://github.com/TuSimple/mxnet-operator/blob/master/README.md).


