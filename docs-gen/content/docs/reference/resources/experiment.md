---
title: "Experiment"
weight: 4
---

The Experiment resource represents the lifecycle of Experiments,
and can be configured as follows:

```yaml
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Experiment
metadata:
  name: penguin-experiment
spec:
  provider: kfp
  description: 'An experiment for the penguin pipeline'
```

## Fields

| Name               | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `spec.provider`    | The name of the associated [Provider resource](../provider/). |
| `spec.description` | The description of the experiment.                            |
