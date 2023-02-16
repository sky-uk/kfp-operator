---
title: "Run Completion Events"
weight: 3
---

![Model Serving](/images/run-completion.png)

The KFP-Operator Events system provides a [NATS Event bus](https://nats.io/) in the operator namespace to consume events from. 
To use it, users can to create an Argo-Events [NATS Eventsource](https://argoproj.github.io/argo-events/eventsources/setup/nats/) as follows:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: run-completion
spec:
  nats:
    run-completion:
      connectionBackoff:
        duration: 10s
        factor: 2
        jitter: 0.2
        steps: 5
      jsonBody: true
      subject: events
      url: nats://eventbus-kfp-operator-events-stan-svc.kfp-operator.svc:4222
```

The specification of the events follows [CloudEvents](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/formats/json-format.md):

```json
{
  "id": "{{ UNIQUE_MESSAGE_ID }}",
  "specversion": "1.0",
  "source": "{{ PROVIDER_NAME }}",
  "type": "org.kubeflow.pipelines.run-completion",
  "datacontenttype": "application/json",
  "data": {
    "status": "succeeded|failed",
    "pipelineName":"{{ PIPELINE_NAME }}",
    "servingModelArtifacts": [
      {
        "name":"{{ PIPELINE_NAME }}:{{ WORKFLOW_NAME }}:Pusher:pushed_model:{{ PUSHER_INDEX }}",
        "location":"gs://{{ PIPELINE_ROOT }}/Pusher/pushed_model/{{ MODEL_VERSION }}"
      }
    ]
  }
}
```

A sensor for the pipeline `penguin-pipeline` could look as follows:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Sensor
metadata:
  name: penguin-pipeline-model-update
spec:
  dependencies:
    - name: run-completion
      eventSourceName: run-completion
      eventName: run-completion
      filters:
        data:
          - path: body.status
            type: string
            comparator: "="
            value:
              - "succeeded"
          - path: body.pipelineName
            type: string
            comparator: "="
            value:
              - "penguin-pipeline"
  triggers:
    - template:
        name: log
        log: {}
```

For more information and an in-depth example, see the [Quickstart Guide](../../getting-started#5-optional-deploy-newly-trained-models) and [Argo-Events Documentation](https://argoproj.github.io/argo-events/).

Please make sure to provide an event bus for the eventsource and the sensor to connect to.
You can define a default event bus, which does not require further configuration on either end, as follows:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: EventBus
metadata:
  name: default
spec:
  nats:
    native: {}
```
