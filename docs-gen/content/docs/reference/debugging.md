---
title: "Debugging"
weight: 3
---

## Debugging Configuration

Debugging options can be set:
 - in the `debug` section of operator's configuration
 - serialised as JSON in the `pipelines.kubeflow.org/debug` annotation of the managed resources

| Option | Description | Example |
| --- | --- | --- |
| `keepWorkflows` | Don't delete workflows after state transitions. This flag is useful for debugging workflows that the operator has created. | `true`,`false` |

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
apiVersion: pipelines.kubeflow.org/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-sample
  annotations:
    pipelines.kubeflow.org/debug: '{ "a": true, "b": false }'
spec:
  ...
```

The above configuration would result in both options `a` and `b` being active.

## Kubernetes Events

The operator emits Kubernetes events for all resource transitions which can be viewed using `kubectl describe`.

Example:

```shell 
$ kubectl describe pipeline pipeline-sample
...
Events:
  Type     Reason      Age    From          Message
  ----     ------      ----   ----          -------
  Normal   Syncing     5m54s  kfp-operator  Updating [version: "v5-841641"]
  Warning  SyncFailed  101s   kfp-operator  Failed [version: "v5-841641"]: pipeline update failed
  Normal   Syncing     9m47s  kfp-operator  Updating [version: "57be7f4-681dd8"]
  Normal   Synced      78s    kfp-operator  Succeeded [version: "57be7f4-681dd8"]
```
