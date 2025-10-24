---
title: "Kubeflow-Pipelines K8s Operator"
linkTitle: "KFP-Operator"
---

<!-- Hero section -->
{{< blocks/cover title="Kubeflow Pipelines Operator">}}
<div class="mx-auto">
  <p class="lead mt-5">Manage ML pipelines as Kubernetes resources using GitOps and declarative configuration.</p>
  <a
    class="btn btn-lg btn-primary mr-3 mb-4"
    href="versions/v0.7.0"
  >
    Get Started <i class="fas fa-arrow-alt-circle-right ml-2"></i>
  </a>
  <a
    class="btn btn-lg btn-secondary mr-3 mb-4"
    href="{{< param "github_project_repo" >}}"
  >
    See the Code <i class="fab fa-github ml-2"></i>
  </a>
</div>

{{< /blocks/cover >}}

{{% blocks/lead type="row" color="white" %}}
The KFP Operator provides a **Kubernetes-native API** for [Kubeflow Pipelines](https://www.kubeflow.org/docs/components/pipelines/).
Deploy and manage ML pipelines using kubectl, Helm, and GitOps workflows with [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).
{{% /blocks/lead %}}

{{% blocks/section type="row" color="secondary" %}}
{{% blocks/feature icon="fa-cube" title="Kubernetes-Native" %}}
Manage ML pipelines as Kubernetes resources using kubectl, Helm, and GitOps workflows.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-sync-alt" title="Event-Driven" %}}
Trigger pipeline runs automatically based on schedules, events, or data changes.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-shield-alt" title="Production-Ready" %}}
Enterprise security, observability, and integration with existing Kubernetes infrastructure.
{{% /blocks/feature %}}

{{% /blocks/section %}}



{{< blocks/section color="white" >}}
<div class="col-12">
<h2 class="text-center">How It Works</h2>
</div>

<div class="col-lg-6">
<h3>Define Pipelines as Code</h3>
<p>Create Kubernetes manifests for your ML pipelines and version control them with your code.</p>

<pre><code>
apiVersion: pipelines.kubeflow.org/v1beta1
kind: Pipeline
metadata:
  name: training-pipeline
spec:
  image: my-org/ml-pipeline:v1.2.0
  env:
  - name: MODEL_VERSION
    value: "v2.1"
</code></pre>

</div>

<div class="col-lg-6">
<h3>Deploy with kubectl</h3>
<p>Use standard Kubernetes tools to deploy and manage your ML workflows.</p>

<pre><code>
# Deploy your pipeline
kubectl apply -f pipeline.yaml

# Trigger a run
kubectl apply -f runconfiguration.yaml

# Monitor status
kubectl get mlr,mlrc
</code></pre>
</div>

{{< /blocks/section >}}

{{< blocks/section color="secondary" >}}
<div class="col-12 text-center">
<h2>Get Started</h2>

<div class="row justify-content-center mt-4">
    <div class="col-lg-12">
    <div class="row">
    <div class="col-md-4 p-1">
        <a href="versions/v0.7.0/getting-started/overview/" class="btn btn-lg btn-block btn-primary">
            <strong>Quick Start</strong><br>
            <small>Installation & first pipeline</small>
        </a>
    </div>
    <div class="col-md-4 p-1">
        <a href="versions/v0.7.0/examples/" class="btn btn-lg btn-block btn-primary">
            <strong>Examples</strong><br>
            <small>Sample pipelines & use cases</small>
        </a>
    </div>
    <div class="col-md-4 p-1">
        <a href="versions/v0.7.0/reference/" class="btn btn-lg btn-block btn-primary">
            <strong>API Reference</strong><br>
            <small>Complete resource specs</small>
        </a>
    </div>
    </div>
    </div>
</div>

<div class="mt-4">
<h4>Installation</h4>
<pre><code>helm repo add kfp-operator https://sky-uk.github.io/kfp-operator/
helm install kfp-operator kfp-operator/kfp-operator</code></pre>
</div>
</div>
{{< /blocks/section >}}

{{< blocks/section color="white" type="row" >}}
{{% blocks/feature icon="fab fa-github" title="Open Source" %}}
100% open source and welcomes contributions. Built by Sky's ML Platform team and used in production.

[**View on GitHub →**]({{< param "github_project_repo" >}})
{{% /blocks/feature %}}

{{% blocks/feature icon="fas fa-comments" title="Community" %}}
Get help and connect with other users in GitHub Discussions.

[**Join Discussions →**](https://github.com/sky-uk/kfp-operator/discussions)
{{% /blocks/feature %}}

{{% blocks/feature icon="fas fa-bug" title="Issues & Feedback" %}}
Report bugs and request features on GitHub Issues.

[**Report Issues →**](https://github.com/sky-uk/kfp-operator/issues)
{{% /blocks/feature %}}
{{< /blocks/section >}}

