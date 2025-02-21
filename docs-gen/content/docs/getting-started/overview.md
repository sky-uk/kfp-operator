---
title: "Overview"
weight: 1
---

The Kubeflow Pipelines Operator (KFP Operator) offers a declarative API designed to streamline the management and 
execution of ML pipelines using Resource Definitions across various providers. 
A "provider" refers to a runtime environment that handles the orchestration and execution of these pipelines and 
associated resources.

### Why KFP Operator

This project was initiated with the goal of **promoting best practices in Machine Learning engineering** while minimizing 
the operational complexities involved in deploying, executing, and maintaining training pipelines. This project seeks to
move away from manual, error-prone, copy-and-paste deployments, and towards a **declarative, transparent, and 
self-service model**.

By configuring simple Kubernetes resources, machine learning practitioners can run their desired training pipelines 
in each environment on the path to production in a repeatable, testable and scalable way. When linked with serving 
components, this provides a fully testable path to production for machine learning systems.

![cd-ct]({{< param "subpath" >}}/master/images/cd-ct.svg)

The KFP Operator bridges the gap between Continuous Delivery (CD) and Continuous Training (CT), enabling Level 2 of the
[MLOps Maturity model](https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines-in-machine-learning#mlops_level_2_cicd_pipeline_automation).

![mlops maturity level]({{< param "subpath" >}}/master/images/mlops-maturity.svg)
