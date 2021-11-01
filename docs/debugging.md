# Debugging

Debugging options can be set:
 - in the `debug` section of operator's configuration
 - serialised as JSON in the `pipelines.kubeflow.org/debug` annotation of the managed resources

| Flag | Description |
| --- | --- |
| `keepWorkflows` | Don't delete workflows after state transitions. This flag is useful for debugging workflows that the operator has created. |

Options set in the operator's configuration act as the lower bound to those defined in the resources.
This means that resources can increase debugging, but never decrease it.

Example:

```yaml
apiVersion: config.kubeflow.org/v1
kind: KfpControllerConfig
metadata:
  name: operatorconfig-sample
spec:
  debug: 
    a: false
    b: true
```

```yaml
apiVersion: pipelines.kubeflow.org/v1
kind: Pipeline
metadata:
  name: pipeline-sample
  annotations:
    pipelines.kubeflow.org/debug: '{ "a": true, "b": false }'
spec:
  ...
```

The above configuration would result in both options `a` and `b` being active.