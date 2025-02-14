---
title: "Using Multiple Providers"
weight: 4
type: docs
---

The KFP operator supports multiple provider backends.

Changing the provider of a resource that was previously managed by another provider will result in the resource erroring.
Any referenced resources must always match the provider of the referencing resource (e.g. RunConfiguration to Pipeline) as updates are not propagated or checked and will result in runtime errors on the provider.
