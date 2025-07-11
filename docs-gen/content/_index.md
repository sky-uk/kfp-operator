---
title: "Kubeflow-Pipelines K8s Operator"
linkTitle: "KFP-Operator"
---

<!-- Hero section -->
{{< blocks/cover title="Kubeflow Pipelines Operator" image_anchor="center" color="dark" >}}
<div class="mx-auto">
  <p class="lead mt-5"></p>
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
{{< blocks/link-down color="info" >}}
{{< /blocks/cover >}}

{{% blocks/lead color="primary" %}}
The Kubeflow Pipelines Operator provides a declarative API for managing and running machine learning pipelines on [Kubeflow](https://www.kubeflow.org) with [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).
{{% /blocks/lead %}}

{{< blocks/section color="dark" type="row">}}
{{< blocks/feature icon="fa-anchor" title="Kubernetes Resources for KFP Entities" >}}
Deploy and manage KFP resources using you favourite tooling.

{{< /blocks/feature >}}

{{< blocks/feature icon="fa-bullhorn" title="Continuous Training Events" >}}
React to machine learning pipeline training runs declaratively.

{{< /blocks/feature >}}

{{< blocks/feature icon="fa-puzzle-piece" title="Unintrusive" >}}
Uses the same great technologies like KFP itself.

{{< /blocks/feature >}}

{{< /blocks/section >}}
