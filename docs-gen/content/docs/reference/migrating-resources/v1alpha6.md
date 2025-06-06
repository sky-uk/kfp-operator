---
title: "v1alpha5 to v1alpha6"
weight: 1
---

This guide documents how to upgrade `pipelines.kubeflow.org` resources from `v1alpha5` to `v1alpha6`.

Follow the steps below for every `Pipeline`, `RunConfiguration`, `Experiment`, and `Provider` you have deployed.

## Pipeline
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.provider` to the value of the `pipelines.kubeflow.org/provider` annotation in `metadata.annotations`.
3. Remove the `pipelines.kubeflow.org/provider` annotation from `metadata.annotations`.

### Example
The example below shows the required changes for migrating a Pipeline resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Pipeline
metadata:
  name: my-training-pipeline
  namespace: my-namespace
- annotations:      
-   pipelines.kubeflow.org/provider: vai
spec:
+ provider: vai
  image: registry/mypipelineimage
  tfxComponents: pipeline.create_components
  beamArgs:
    - name: anArg
      value: aValue
```

---

## RunConfiguration

1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.provider` to the value of the `pipelines.kubeflow.org/provider` annotation in `metadata.annotations`.
3. Remove the `pipelines.kubeflow.org/provider` annotation from `metadata.annotations`.
4. Change the `spec.trigger.schedules` block from being a list of cron expression strings to a list of objects, where each object contains `cronExpression`. If required, users can set the `startTime`, and `endTime` to define when the schedule should start or stop.

### Example
The example below shows the required changes for migrating a RunConfiguration resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: RunConfiguration
metadata:
  name: my-run-config
  namespace: my-namespace
- annotations:      
-   pipelines.kubeflow.org/provider: vai
spec:
  run: 
+   provider: vai
    pipeline: my-training-pipeline
  triggers:
    schedules:
-     - 0 * * * *
+     - cronExpression: 0 * * * *
```

---

## Experiment
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.provider` to the value of the `pipelines.kubeflow.org/provider` annotation in `metadata.annotations`.
3. Remove the `pipelines.kubeflow.org/provider` annotation from `metadata.annotations`.

### Example
The example below shows the required changes for migrating an Experiment resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Experiment
metadata:
  name: my-experiment
  namespace: my-namespace
- annotations:      
-   pipelines.kubeflow.org/provider: vai
spec:
+ provider: vai
  description: 'An experiment'
```

---

## Provider
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.serviceImage` to the relevant [provider service](../../reference/providers/overview.md) image tag.

### Example
The example below shows the required changes for migrating a Provider resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Provider
metadata:
  name: vai
  namespace: my-provider-namespace
spec:
+ serviceImage: kfp-operator-kfp-provider-service:<version>
  image: kfp-operator-vai-provider:<version>
  defaultBeamArgs:
  - name: project
    value: <project>
  executionMode: v2
  pipelineRootStorage: gs://<storage_location>
  serviceAccount: kfp-operator-vai
  parameters:
    eventsourcePipelineEventsSubscription: kfp-operator-vai-run-events-eventsource
    maxConcurrentRunCount: 1
    pipelineBucket: pipeline-storage-bucket
    vaiJobServiceAccount: kfp-operator-vai@<project>.iam.gserviceaccount.com
    vaiLocation: europe-west2
    vaiProject: <project>
```
