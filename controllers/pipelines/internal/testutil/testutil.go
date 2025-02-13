//go:build decoupled || integration || unit

package testutil

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	K8sClient  client.Client
	Ctx        context.Context
	TestConfig config.KfpControllerConfigSpec
	Provider   *pipelinesv1.Provider
)

func SchemeWithCrds() *runtime.Scheme {
	scheme := runtime.NewScheme()

	scheme.AddKnownTypes(pipelinesv1.GroupVersion, &pipelinesv1.RunConfiguration{}, &pipelinesv1.Run{}, &pipelinesv1.Provider{})
	scheme.AddKnownTypes(pipelinesv1.GroupVersion, &metav1.Status{})

	metav1.AddToGroupVersion(scheme, pipelinesv1.GroupVersion)
	return scheme
}
