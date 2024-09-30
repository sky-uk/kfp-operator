//go:build decoupled || integration

package pipelines

import (
	"context"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	k8sClient  client.Client
	ctx        context.Context
	testConfig config.KfpControllerConfigSpec
)
