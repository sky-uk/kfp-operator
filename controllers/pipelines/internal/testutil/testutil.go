//go:build decoupled || integration || unit

package testutil

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

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

func SchemeWithCRDs() *runtime.Scheme {
	scheme := runtime.NewScheme()

	groupVersion := schema.GroupVersion{Group: "pipelines.kubeflow.org", Version: "v1alpha6"}
	scheme.AddKnownTypes(groupVersion, &pipelinesv1.RunConfiguration{}, &pipelinesv1.Run{}, &pipelinesv1.Provider{})
	scheme.AddKnownTypes(groupVersion, &metav1.Status{})

	metav1.AddToGroupVersion(scheme, groupVersion)
	return scheme
}
