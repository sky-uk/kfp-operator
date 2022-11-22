---
title: "Debugging"
weight: 3
---

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
