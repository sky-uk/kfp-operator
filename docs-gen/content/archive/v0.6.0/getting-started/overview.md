---
title: "Overview"
weight: 1
type: docs
---

The Kubeflow Pipelines Operator (KFP Operator) provides a declarative API for managing and running ML pipelines with Resource Definitions on multiple providers.
A provider is a runtime environment for managing and executing ML pipelines and related resources.

### Why KFP Operator

We started this project to promote the best engineering practices in the Machine Learning process, while reducing the operational overhead associated with deploying, running and maintaining training pipelines. We wanted to move away from a manual, opaque, copy-and-paste style deployment and closer to a declarative, traceable, and self-serve approach.


By configuring simple Kubernetes resources, machine learning practitioners can run their desired training pipelines in each environment on the path to production in a repeatable, testable and scalable way. When linked with serving components, this provides a fully testable path to production for machine learning systems.

![cd-ct]({{< param "subpath" >}}/images/cd-ct.svg)

Through separating training code from infrastructure, KFP Operator provides 
the link between CD and CT to provide Level 2 of the [MLOps Maturity model](https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines-in-machine-learning#mlops_level_2_cicd_pipeline_automation). 

![mlops maturity level]({{< param "subpath" >}}/images/mlops-maturity.svg)
