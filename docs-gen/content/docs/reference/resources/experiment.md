---
title: "Experiment"
weight: 4
---

The Experiment resource represents the lifecycle of Experiments,
and can be configured as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1alpha5
kind: Experiment
metadata:
  name: penguin-experiment
spec:
  description: 'An experiment for the penguin pipeline'
```

## Fields

| Name               | Description                        |
|--------------------|------------------------------------|
| `spec.description` | The description of the experiment. |
