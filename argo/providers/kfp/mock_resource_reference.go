//go:build decoupled || unit

package kfp

import (
	"context"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type MockK8sResources struct {
	k8sClient          *dynamic.Interface
	resourceReferences *ResourceReferences
}

func (mock MockK8sResources) GetUnderlyingClient() dynamic.Interface {
	if mock.k8sClient != nil {
		return *mock.k8sClient
	}
	panic("not expected to be called in test")
}

func (mock MockK8sResources) GetRunArtifactDefinitions(_ context.Context, namespacedName types.NamespacedName, gvr schema.GroupVersionResource) ([]pipelinesv1.OutputArtifact, error) {
	Expect(namespacedName.Name).To(Equal(mock.resourceReferences.RunConfigurationName.Name))
	Expect(namespacedName.Namespace).To(Equal(mock.resourceReferences.RunConfigurationName.Namespace))
	Expect(gvr).To(Equal(base.RunConfigurationGVR))
	return nil, nil
}
