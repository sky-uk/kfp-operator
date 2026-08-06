| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `containerRegistry` | Container Registry base path for all container images | `"ghcr.io/kfp-operator"` |
| `crds.create` | Install the operator's Custom Resource Definitions | `true` |
| `crds.keep` | Retain the Custom Resource Definitions when the chart is uninstalled | `true` |
| `fullnameOverride` | Override the fully qualified name of chart resources | `""` |
| `kfp-provider-workflows.enabled` | Enable the kfp-provider-workflows subchart. Disabled by default; install the subchart separately, one release per provider namespace | `false` |
| `logging.verbosity` | Logging verbosity for all components - see the [logging documentation]({{< param "github_project_repo" >}}/blob/master/CONTRIBUTING.md#logging) for valid values | `nil` |
| `manager.configuration` | Manager configuration as defined in [Configuration](../configuration/operator-configuration) (note that you can omit `compilerImage` and `kfpSdkImage` when specifying `containerRegistry` as default values will be applied) | `{}` |
| `manager.leaderElection.enabled` | Toggle leader election - defaults to `true` | `true` |
| `manager.leaderElection.id` | Leader election Lease resource name - defaults to `kfp-operator-lock` | `"kfp-operator-lock"` |
| `manager.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the manager's pods | `{}` |
| `manager.monitoring.create` | Create the manager's monitoring resources | `false` |
| `manager.monitoring.rbacSecured` | Enable additional RBAC-based security | `false` |
| `manager.monitoring.serviceMonitor.create` | Create a ServiceMonitor for the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) | `false` |
| `manager.monitoring.serviceMonitor.endpointConfiguration` | Additional configuration to be used in the service monitor endpoint (path, port and scheme are provided) | `{}` |
| `manager.multiversion.enabled` | Enable multiversion API. Should be used in production to allow version migration, disable for simplified installation | `false` |
| `manager.multiversion.storedVersion` | Specifies which CRD version should be set as the stored version. Only takes effect if `manager.multiversion.enabled` is set to `true`. Defaults to the latest version. | `"v1beta1"` |
| `manager.rbac.create` | Create roles and rolebindings for the operator | `true` |
| `manager.replicas` | Number of replicas for the manager deployment | `1` |
| `manager.resources` | Manager resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources) | `{"limits":{"cpu":"100m","memory":"300Mi"},"requests":{"cpu":"100m","memory":"200Mi"}}` |
| `manager.runcompletionWebhook.endpoints` | Array of endpoints for the run completion event handlers to be called when a run completion event is passed | `[]` |
| `manager.runcompletionWebhook.servicePort` | Port for the run completion event webhook service to listen on - defaults to 8082 | `8082` |
| `manager.serviceAccount.create` | Create the manager's service account or expect it to be created externally | `true` |
| `manager.serviceAccount.name` | Manager service account's name | `"kfp-operator-controller-manager"` |
| `manager.webhookCertificates.caBundle` | CA bundle of the certificate authority that has signed the webhook's certificate, required if the `custom` provider is chosen | `""` |
| `manager.webhookCertificates.provider` | K8s conversion webhook TLS certificate provider - choose `cert-manager` for Helm to deploy certificates if cert-manager is available or `custom` otherwise (see below) | `"cert-manager"` |
| `manager.webhookCertificates.secretName` | Name of a K8s secret deployed into the operator namespace to secure the webhook endpoint with, required if the `custom` provider is chosen | `""` |
| `manager.webhookServicePort` | Port for the webhook service to listen on - defaults to 9443 | `9443` |
| `namespace.create` | Create the namespace for the operator | `true` |
| `namespace.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the operator namespace | `{}` |
| `namespace.name` | Operator namespace name | `"kfp-operator-system"` |
| `provider.env` | Additional environment variables for provider containers | `[]` |
| `provider.labels` | Additional labels applied to provider resources | `{}` |
| `provider.metricsPort` | Port for provider metrics endpoints | `8081` |
| `provider.podTemplateLabels` | Additional labels applied to provider pod templates | `{}` |
| `provider.replicas` | Number of replicas for provider deployments | `1` |
| `provider.resources` | Provider resources as per [k8s documentation](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources) | `{"limits":{"cpu":"500m","memory":"256Mi"},"requests":{"cpu":"250m","memory":"128Mi"}}` |
| `provider.servicePort` | Port for provider services to listen on | `8080` |
| `provider.volumeMounts` | Additional volume mounts for provider containers | `[]` |
| `provider.volumes` | Additional volumes to mount into provider pods | `[]` |
| `runcompletionEventTrigger.enabled` | Whether the run completion event trigger should be installed - defaults to `false` | `false` |
| `runcompletionEventTrigger.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the run completion event trigger's pods | `{}` |
| `runcompletionEventTrigger.metrics.port` | Port for the run completion event trigger metrics endpoint | `8081` |
| `runcompletionEventTrigger.monitoring` | Whether monitoring resources should be created for the run completion event trigger - defaults to `false` | `false` |
| `runcompletionEventTrigger.nats.server.port` | Port of the NATS server the run completion event trigger connects to | `4222` |
| `runcompletionEventTrigger.nats.subject` | NATS subject the run completion event trigger publishes events to | `"events"` |
| `runcompletionEventTrigger.replicas` | Number of replicas for the run completion event trigger deployment | `1` |
| `runcompletionEventTrigger.server.port` | Port for the run completion event trigger gRPC server to listen on | `50051` |
| `statusFeedback.enabled` | Whether run completion eventing and status update feedback loop should be installed - defaults to `false` | `false` |
