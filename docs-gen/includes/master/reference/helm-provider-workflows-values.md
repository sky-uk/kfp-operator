| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `argo.containerDefaults` | Container spec defaults applied to the Argo workflow pods | `{}` |
| `argo.metadata` | Metadata (labels/annotations) applied to the Argo workflow pods | `{}` |
| `argo.rbac.create` | Create the namespace-scoped `workflow-executor` `Role` and its `RoleBinding`. Set to `false` to manage these resources externally | `true` |
| `argo.securityContext` | [Security Context](https://argo-workflows.readthedocs.io/en/latest/workflow-pod-security-context/) applied to the Argo workflow pods. Set to `null` or `runAsNonRoot` to `false` to run as root | `{"fsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` |
| `argo.serviceAccount.create` | Create the Argo workflow `ServiceAccount`. Set to `false` to reuse an existing one | `true` |
| `argo.serviceAccount.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the Argo workflow `ServiceAccount` | `{}` |
| `argo.serviceAccount.name` | `ServiceAccount` the Argo workflows run as, created in the provider namespace | `"kfp-operator-argo"` |
| `argo.stepTimeoutSeconds.compile` | Timeout in seconds for compiler steps | `1800` |
| `argo.stepTimeoutSeconds.default` | Default timeout in seconds for workflow steps | `300` |
| `argo.ttlStrategy.secondsAfterCompletion` | [TTL Strategy](https://argoproj.github.io/argo-workflows/fields/#ttlstrategy) - seconds to retain completed Argo Workflows before cleanup | `3600` |
| `namespace` | Namespace the `WorkflowTemplate`s, RBAC and `Provider` resource are created in. Must match the namespace the `Provider` runs in. Defaults to the release namespace when left empty | `""` |
| `provider.allowedNamespaces` | Namespaces allowed to reference this provider. An empty list allows all namespaces | `[]` |
| `provider.create` | When `true`, the chart also renders the `Provider` custom resource | `true` |
| `provider.frameworks` | Pipeline frameworks the provider supports (see the [Provider CRD reference](../../reference/resources/provider)) | `[]` |
| `provider.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the `Provider` resource | `{}` |
| `provider.name` | `Provider` resource name. Defaults to the release name when left empty | `""` |
| `provider.parameters` | Free-form provider parameters (map of name to value) | `{}` |
| `provider.pipelineRootStorage` | Pipeline root storage location. Required when `provider.create` is `true` | `""` |
| `provider.serviceAccount.create` | Create the provider service `ServiceAccount`. Set to `false` to reuse an existing one | `true` |
| `provider.serviceAccount.metadata` | [Object Metadata](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/object-meta/#ObjectMeta) for the provider service `ServiceAccount` | `{}` |
| `provider.serviceAccount.name` | Provider service `ServiceAccount` name, created in the provider namespace and referenced by the `Provider` spec. Defaults to `kfp-provider-<provider-name>` when left empty. Its cluster-scoped viewer/eventing bindings are managed externally by the platform | `""` |
| `provider.serviceImage` | Provider service image. Required when `provider.create` is `true` | `""` |
