---
title: "v1alpha5 to v1alpha6"
weight: 1
---

This guide documents how to upgrade `pipelines.kubeflow.org` resources from `v1alpha5` to `v1alpha6`.

Follow the steps below for every `Pipeline`, `RunConfiguration`, `Experiment`, `Run`, `RunSchedule`, and `Provider` you deploy yourself.

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
4. Change the `spec.trigger.schedules` block from being a list of cron expression strings to a list of objects, where each object contains `cronExpression`. If required, users can set the `startTime` and `endTime` fields on the same object to define when the schedule should start or stop.

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
    runtimeParameters:
    - name: TRAINING_RUNS
      value: '100'
  triggers:
    schedules:
-   - 0 * * * *
+   - cronExpression: 0 * * * *
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

## Run
> In general, we expect users to deploy [RunConfigurations](../../runconfiguration) to configure the lifecycle of their runs, leaving the management of `Runs` to the operator. However, if users are deploying `Runs` themselves, they can follow the below steps to migrate the resource version.
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.provider` to the value of the `pipelines.kubeflow.org/provider` annotation in `metadata.annotations`.
3. Remove the `pipelines.kubeflow.org/provider` annotation from `metadata.annotations`.

### Example
The example below shows the required changes for migrating a Run resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: Run
metadata:
  generateName: penguin-pipeline-run-
- annotations:      
-   pipelines.kubeflow.org/provider: vai
spec:
+ provider: vai
  pipeline: penguin-pipeline
  experimentName: penguin-experiment
  runtimeParameters:
  - name: TRAINING_RUNS
    value: '100'
  - name: EXAMPLES
    valueFrom:
      runConfigurationRef:
        name: base-namespace/penguin-pipeline-example-generator-runconfiguration
        outputArtifact: examples
  artifacts:
  - name: serving-model
    path: 'Pusher:pushed_model:0[pushed == 1]'
```

---

## RunSchedule
> In general, we expect users to deploy [RunConfigurations](../../runconfiguration) to configure the lifecycle of their runs, leaving the management of `RunSchedules` to the operator. However, if users are deploying `RunSchedules` themselves, they can follow the below steps to migrate the resource version.
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.provider` to the value of the `pipelines.kubeflow.org/provider` annotation in `metadata.annotations`.
3. Remove the `pipelines.kubeflow.org/provider` annotation from `metadata.annotations`.
4. Change `spec.schedule` from being a cron expression string to an object which contains `cronExpression`. If required, users can also set the `startTime` and `endTime` fields on the same object to define when the schedule should start or stop.

### Example
The example below shows the required changes for migrating a RunSchedule resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha5
+ apiVersion: pipelines.kubeflow.org/v1alpha6
kind: RunSchedule
metadata:
  generateName: penguin-pipeline-run-
- annotations:      
-   pipelines.kubeflow.org/provider: vai
spec:
+ provider: vai
  pipeline: penguin-pipeline
  experimentName: penguin-experiment
 runtimeParameters:
  - name: TRAINING_RUNS
    value: '100'
  - name: EXAMPLES
    valueFrom:
      runConfigurationRef:
        name: base-namespace/penguin-pipeline-example-generator-runconfiguration
        outputArtifact: examples
  artifacts:
  - name: serving-model
    path: 'Pusher:pushed_model:0[pushed == 1]'
  schedule:
    cronExpression: 0 * * * *
```

---

## Provider
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha5` to `pipelines.kubeflow.org/v1alpha6`.
2. Set `spec.serviceImage` to the relevant [provider service](../../../providers/overview) image tag.

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
