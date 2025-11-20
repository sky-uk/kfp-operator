//go:build decoupled || integration || unit

package testutil

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/sky-uk/kfp-operator/internal/config"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	K8sClient  client.Client
	Ctx        context.Context
	TestConfig config.ConfigSpec
	Provider   *pipelineshub.Provider
)

func SchemeWithCrds() *runtime.Scheme {
	scheme := runtime.NewScheme()

	scheme.AddKnownTypes(pipelineshub.GroupVersion, &pipelineshub.RunConfiguration{}, &pipelineshub.Run{}, &pipelineshub.Provider{})
	scheme.AddKnownTypes(pipelineshub.GroupVersion, &metav1.Status{})

	metav1.AddToGroupVersion(scheme, pipelineshub.GroupVersion)
	return scheme
}
