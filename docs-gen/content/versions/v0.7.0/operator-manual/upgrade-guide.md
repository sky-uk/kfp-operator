---
title: "Upgrade Guide"
weight: 1
---

## Upgrading to a New CRD Version

Upgrading the KFP Operator to a new CRD version is quite simple. These are the steps you need to follow:

1) Make sure the `manager.multiversion.storedVersion` field in your Helm values (`values.yaml` file) is set to the new version, e.g. `manager.multiversion.storedVersion: v1beta1` (see [Configuration](../../reference/configuration) for a full list of configuration options). This field will always be set to the latest CRD version in the default Helm values, so you can omit this step if you're not overriding the default values with your own `values.yaml` file.
1) Install/upgrade the operator in the desired cluster as normal. 

## Upgrading to an Unstable CRD Version

If you want to upgrade the KFP Operator to a CRD version that hasn't been marked as stable yet (e.g. from the latest commit on `master` instead of a stable GitHub release) and you want to ensure that rolling back to a stable version is easy, you can override the default stored version to be the currently stable version:

1) Set the `manager.multiversion.storedVersion` field in your Helm values file to the currently stable version, e.g. `manager.multiversion.storedVersion: v1alpha6`. This ensures that the new version will not become the stored version, which means K8s will not store resources in that version, making it easy to roll back.
1) Install/upgrade the operator in the desired cluster as normal. A couple of things to keep in mind:
    - Even though the new version will not be stored, K8s will still be able to serve it and clients like `kubectl` will, in fact, get this new version by default (in accordance with the [version priority algorithm](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#version-priority)). This means that when users retrieve their resources, K8s will convert them to the new version, even if they were created with the old one. This shouldn't cause any problems because the conversion webhooks can handle converting resources between any two versions.
    - While it is technically possible to manually set `served: false` on the new version in all CRDs to hide it from users until it's ready to be used, this is not an option in practice. This is because the operator will always request the hub (latest) version of any resource it needs to reconcile, and if K8s can't serve that version then the controller manager will error and crash.
1) The operator can continue to be incrementally upgraded with new changes on `master` by following these steps.
1) Once the new version has become stable, you can follow the steps in [Upgrading to a New CRD Version](#upgrading-to-a-new-crd-version) to set the new version as the stored version.

As long as a version has never been set as the stored version, you can always roll back any commit that's part of that version (including the commit that introduced the version in the first place) by simply releasing the operator from the version that is 1 commit behind the commit you wish to roll back.
