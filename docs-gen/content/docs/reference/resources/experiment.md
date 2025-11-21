---
title: "Experiment"
weight: 5
---

The Experiment resource represents the lifecycle of Experiments,
and can be configured as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Experiment
metadata:
  name: penguin-experiment
spec:
  provider: provider-namespace/provider-name
  description: 'An experiment for the penguin pipeline'
```

## Fields

| Name               | Description                                                                                                                             |
|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| `spec.provider`    | The namespace and name of the associated [Provider resource](../provider/) separated by a `/`, e.g. `provider-namespace/provider-name`. |
| `spec.description` | The description of the experiment.                                                                                                      |
