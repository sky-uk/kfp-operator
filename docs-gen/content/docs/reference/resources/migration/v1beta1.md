---
title: "v1alpha6 to v1beta1"
weight: 2
---

This guide documents how to upgrade `pipelines.kubeflow.org` resources from `v1alpha6` to `v1beta1`. 

Follow the steps below for every `Pipeline`, `RunConfiguration`, `Experiment`, `Run`, `RunSchedule`, and `Provider` you deploy yourself.

## Pipeline
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha6` to `pipelines.kubeflow.org/v1beta1`.
2. Ensure that the `spec.provider` field includes the namespace that the Provider resource is deployed in.
3. Add `spec.frameworks`, which is an object with `name` and `parameters` fields. `parameters` are framework specific parameters, such as `components` and `beamArgs`. To use the `tfx` framework that was the only option under versions `v1alpha6` and below, set the `name` field to `tfx`, add the path to the function that returns the tfx components under `spec.frameworks.components`, and add any required beamArgs like the example below.
4. Remove `spec.tfxComponents`.
5. Remove `spec.beamArgs`.

### Example
The example below shows the required changes for migrating a Pipeline resource from `v1alpha6` to `v1beta1`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: my-training-pipeline
  namespace: my-namespace
spec:
- provider: vai
+ provider: my-provider-namespace/vai
  image: registry/mypipelineimage
- tfxComponents: pipeline.create_components
+ frameworks:
+   name: tfx
+   parameters:
+     components: base_pipeline.create_components
+     beamArgs:
+     - name: anArg
+       value: aValue
- beamArgs:
- - name: anArg
-   value: aValue
```


## RunConfiguration
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha6` to `pipelines.kubeflow.org/v1beta1`.
2. Ensure that the `spec.run.provider` field includes the namespace that the Provider resource is deployed in.
3. Change the `spec.run.runtimeParameters` field to `spec.run.parameters` if set.

### Example
The example below shows the required changes for migrating a RunConfiguration resource from `v1alpha6` to `v1beta1`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: RunConfiguration
metadata:
  name: my-run-config
  namespace: my-namespace
spec:
  run: 
-   provider: vai
+   provider: my-provider-namespace/vai
    pipeline: my-training-pipeline
-   runtimeParameters:
+   parameters:
    - name: TRAINING_RUNS
      value: '100'
  triggers:
    schedules:
      - cronExpression: 0 * * * *
```

## Experiment
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha6` to `pipelines.kubeflow.org/v1beta1`.
2. Ensure that the `spec.provider` field includes the namespace that the Provider resource is deployed in.

### Example
The example below shows the required changes for migrating an Experiment resource from `v1alpha6` to `v1beta1`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: Experiment
metadata:
  name: my-experiment
  namespace: my-namespace
spec:
- provider: vai
+ provider: my-provider-namespace/vai
  description: 'An experiment'
```

## Run
> In general, we expect users to deploy [RunConfigurations](../../runconfiguration) to configure the lifecycle of their runs, leaving the management of `Runs` to the operator. However, if users are deploying `Runs` themselves, they can follow the below steps to migrate the resource version.
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha6` to `pipelines.kubeflow.org/v1beta1`.
2. Ensure that the `spec.provider` field includes the namespace that the Provider resource is deployed in.
3. Change the `spec.runtimeParameters` field to `spec.parameters` if set.

### Example
The example below shows the required changes for migrating a Run resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: Run
metadata:
  generateName: penguin-pipeline-run-
spec:
- provider: vai
+ provider: my-provider-namespace/vai
  pipeline: penguin-pipeline:v1-abcdef
  experimentName: penguin-experiment
- runtimeParameters:
+ parameters:
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
3. Change the `spec.runtimeParameters` field to `spec.parameters` if set.

### Example
The example below shows the required changes for migrating a RunSchedule resource from `v1alpha5` to `v1alpha6`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: RunSchedule
metadata:
  generateName: penguin-pipeline-run-
spec:
- provider: vai
+ provider: my-provider-namespace/vai
  pipeline: penguin-pipeline:v1-abcdef
  experimentName: penguin-experiment
- runtimeParameters:
+ parameters:
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
1. Change the `apiVersion` from `pipelines.kubeflow.org/v1alpha6` to `pipelines.kubeflow.org/v1beta1`.
2. Add a list of [frameworks](../../../frameworks) supported by the provider in question, including the `name` and `image` of the compiler image. `patches` can be used to perform a patch operation on Pipeline resources, which can be used to provide defaults such as defaultBeamArgs. See the example below and [reference guide](../../../resources/provider) for more detail.
3. If required, add a list of namespaces that resources can reference this provider from under `spec.allowedNamespaces`. See the example below and [reference guide](../../../resources/provider) for more detail.
4. Remove `spec.image` - this was the image of the provider CLI which has now been removed.
5. Remove `spec.defaultBeamArgs`.

### Example
The example below shows the required changes for migrating a Provider resource from `v1alpha6` to `v1beta1`.
```diff
- apiVersion: pipelines.kubeflow.org/v1alpha6
+ apiVersion: pipelines.kubeflow.org/v1beta1
kind: Provider
metadata:
  name: vai
  namespace: my-provider-namespace
spec:
  serviceImage: kfp-operator-kfp-provider-service:<version>
- image: kfp-operator-vai-provider:<version>
- defaultBeamArgs:
- - name: project
-   value: <project>
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
+ frameworks:
+ - name: tfx
+   image: ghcr.io/kfp-operator/kfp-operator-tfx-compiler:version-tag
+   patches:
+   - type: json
+     patch: |
+       [
+         {
+           "op": "add",
+           "path": "/framework/parameters/beamArgs/0",
+           "value": {
+             "name": "project",
+             "value": "<project>"
+           }
+         }
+       ]
+ allowedNamespaces:
+ - default
+ - my-namespace
```
