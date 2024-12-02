//go:build decoupled || integration

package testutil

import (
	"context"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	k8sClient  client.Client
	ctx        context.Context
	TestConfig config.KfpControllerConfigSpec
	Provider   *pipelinesv1.Provider
)
