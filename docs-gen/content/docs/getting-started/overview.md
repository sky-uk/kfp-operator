---
title: "Overview"
weight: 1
---

The Kubeflow Pipelines Operator (KFP Operator) provides a declarative API for managing and running ML pipelines with Resource Definitions on multiple providers.
A provider is a runtime environment for managing and executing ML pipelines and related resources.

### Why KFP Operator

The idea behind the project came about because as engineers we wanted to apply our engineering best practices behind the 
process of deploying a trained model to be served in a production environment. We wanted to get away from the mentality of 
"I just need this trained model in Production" or the "If it's 10 lines of code I'd rather copy and paste it".

Data scientists produce the artifact of the "model training pipeline code" (which is fully promotable through the environments)
and then the deployment configuration for the pipeline etc. can be applied in the different environments, giving the required
training in each environment on the path to production. Linked with the serving component then gives you a fully testable path to production for the model.

![cd-ct](/images/cd-ct.svg)

KFP Operator allows for the separation of the two concerns, data science and deployment configuration. KFP Operator provides 
the link between CD and CT to provide Level 2 of the [MLOps Maturity model](https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines-in-machine-learning#mlops_level_2_cicd_pipeline_automation). 

![mlops maturity level](/images/mlops-maturity.svg)

