---
title: "Configuration"
weight: 1
---

The Kubeflow Pipelines operator can be configured with the following parameters:

| Parameter name       | Description                                                                                                                                                                                                                        | Example                                                |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `defaultBeamArgs`    | Default Beam arguments to which the pipeline-defined ones will be added                                                                                                                                                            | <pre>- name: project<br/>  value: my-gcp-project</pre> |
| `defaultExperiment`  | Default Experiment name to be used for creating pipeline runs                                                                                                                                                                      | `Default`                                              |
| `multiversion`       | If enabled, it will support previous versions of the CRDs, only the latest otherwise                                                                                                                                               | `true`                                                 |
| `pipelineStorage`    | The storage location used by [TFX (`pipeline-root`)](https://www.tensorflow.org/tfx/guide/build_tfx_pipeline) to store pipeline artifacts and outputs - this should be a top-level directory and not specific to a single pipeline | `gcs://kubeflow-pipelines-bucket`                      |
| `providerConfigFile` | File in the controller container containing the provider-specific configuration (see below)                                                                                                                                        | `provider.yaml`                                        |

An example can be found in the [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/controller_manager_config.yaml).

## Provider Configuration

The provider configuration is specific to the implementation. The operator supports the following out of the box.

An example can be found [here](https://github.com/sky-uk/kfp-operator/blob/master/config/manager/provider.yaml).

### Common

| Parameter name  | Description                                                                                                                          | Example                                  |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------|
| `image`         | Container image of the provider                                                                                                      | `kfp-operator-kfp-provider:0.0.2`        |
| `executionMode` | KFP compiler [execution mode](https://kubeflow-pipelines.readthedocs.io/en/latest/source/kfp.dsl.html#kfp.dsl.PipelineExecutionMode) | `v1` (currently KFP) or `v2` (Vertex AI) |

### Kubeflow Pipelines

KFP must be installed in [standalone mode](https://www.kubeflow.org/docs/components/pipelines/installation/standalone-deployment/). Default endpoints are used below.
Optionally, [Argo-Events](https://argoproj.github.io/argo-events/installation/) can be installed for eventing support.

| Parameter name             | Description                                      | Example                                         |
|----------------------------|--------------------------------------------------|-------------------------------------------------|
| `restKfpApiUrl`            | The KFP REST URL available to the operator       | `http://ml-pipeline.kubeflow:8888`              |
| `grpcKfpApiAddress`        | The KFP gRPC address for the eventsource server  | `ml-pipeline.kubeflow-pipelines:8887`           |
| `grpcMetadataStoreAddress` | The MLMD gRPC address for the eventsource server | `metadata-grpc-service.kubeflow-pipelines:8080` |

### Vertex AI Pipelines

![Vertex AI Provider](/images/vai-provider.png)

The following GCP APIs need to be enabled:
- Vertex AI
- Pub/Sub
- Cloud Storage
- Cloud Scheduler

Pub/Sub topics and subscriptions need to be created for:
- Run Intents `provider.configuration.runIntentsTopic`, `provider.configuration.enqueuerRunIntentsSubscription`)
- Runs `provider.configuration.runsTopic`, `provider.configuration.submitterRunsSubscription`

GCS pipeline storage bucket `provider.configuration.pipelineBucket` needs to be created

The following workload-identity-enabled service accounts need to be created with the respective permissions:
- Argo Workflow Runner `manager.argo.serviceAccount`
  - `cloudscheduler.jobs.create`
  - `projects.topics.publish` to the configured Run Intents topic
- Vertex AI Worker `manager.provider.serviceAccount`
  - `projects.subscriptions.pull` from the configured Run Intents and Runs subscriptions
  - `projects.topics.publish` to the configured Runs topic
  - `aiplatform.pipelineJobs.create`
  - `iam.serviceAccounts.actAs` Vertex AI Job Runner
- Vertex AI Job Runner `manager.provider.configuration.vaiJobServiceAccount`
  - all permissions needed by pipeline jobs
- Vertex AI Eventsource Server `eventsource.serviceaccount.name`
  - `projects.subscriptions.pull` from the configured Runs subscriptions
  - `aiplatform.pipelineJobs.get`

[Argo-Events](https://argoproj.github.io/argo-events/installation/) must be installed into the operator's Kubernetes cluster and any namespace that is going to use the eventsource server.

| Parameter name                   | Description                                                   | Example                                                           |
|----------------------------------|---------------------------------------------------------------|-------------------------------------------------------------------|
| `pipelineBucket`                 | GCS bucket where to store the compiled pipeline               | `kfp-operator-pipelines`                                          |
| `vaiProject`                     | Vertex AI GCP project name                                    | `kfp-operator-vertex-ai`                                          |
| `vaiLocation`                    | Vertex AI GCP project location                                | `europe-west2`                                                    |
| `vaiJobServiceAccount`           | Vertex AI GCP service account to run pipeline jobs            | `kfp-operator-vai@kfp-operator-vertex-ai.iam.gserviceaccount.com` |
| `runIntentsTopic`                | Pub/Sub topic name to publish run intents                     | `kfp-operator-run-intents`                                        |
| `enqueuerRunIntentsSubscription` | Subscription on the run intents topic                         | `kfp-operator-runs-enqueuer`                                      |
| `runsTopic`                      | Pub/Sub topic name to publish runs                            | `kfp-operator-runs`                                               |
| `submitterRunsSubscription`      | Subscription on the runs topic for the pipeline job submitter | `kfp-operator-runs-submitter`                                     |
| `eventsourceRunsSubscription`    | Subscription to runs topic for the eventsource server         | `kfp-operator-runs-eventsource`                                   |
