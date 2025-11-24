---
title: "Custom Resources"
linkTitle: "Custom Resources"
description: "Canonical API reference for all Kubernetes Custom Resource Definitions (CRDs)"
weight: 10
---

# Custom Resource Definitions (CRDs)

This section provides the canonical API reference for all Kubernetes Custom Resource Definitions (CRDs) provided by the KFP Operator.

## Resource Overview

The KFP Operator extends Kubernetes with four primary Custom Resource types:

### Resource Hierarchy
```
Provider (Platform Connection)
    ↓
Pipeline (Template Definition)
    ↓
Run (Individual Execution)
    ↑
RunConfiguration (Automation Rules)
```

## Available Resources

### [Pipeline](pipeline/)
Defines reusable ML pipeline templates with container specifications, environment configuration, and execution parameters.

### [Run](run/)
Represents individual pipeline executions with runtime parameters and status tracking.

### [RunConfiguration](runconfiguration/)
Configures automated pipeline execution with scheduling, event triggers, and parameter templates.

### [RunSchedule](runschedule/)
Defines recurring execution schedules for pipelines using cron expressions.

### [Experiment](experiment/)
Defines experiments to group pipelines within kubeflow.

### [Provider](provider/)
Manages connections to ML orchestration platforms with authentication, health monitoring, and load balancing.

## Common Resource Patterns

All Custom Resources share common metadata and status patterns. See individual resource documentation for complete field specifications, validation rules, and examples.
