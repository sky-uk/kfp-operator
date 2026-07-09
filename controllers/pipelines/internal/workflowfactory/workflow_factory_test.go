//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/internal/config"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		owner := pipelineshub.RandomResource()
		provider := pipelineshub.RandomProvider()
		w := ResourceWorkflowFactory[*pipelineshub.TestResource, any]{}
		meta := w.CommonWorkflowMeta(owner, *provider)

		Expect(meta.Namespace).To(Equal(provider.Namespace))
		Expect(meta.GetGenerateName()).To(Equal(owner.GetKind() + "-" + owner.GetName() + "-"))

		Expect(meta.Labels[workflowconstants.OwnerKindLabelKey]).To(Equal(owner.GetKind()))
		Expect(meta.Labels[workflowconstants.OwnerNameLabelKey]).To(Equal(owner.GetName()))
		Expect(meta.Labels[workflowconstants.OwnerNamespaceLabelKey]).To(Equal(owner.GetNamespace()))
	})

	It("uses the provider namespace regardless of config", func() {
		owner := pipelineshub.RandomResource()
		provider := pipelineshub.RandomProvider()
		provider.Namespace = "provider-namespace"
		w := ResourceWorkflowFactory[*pipelineshub.TestResource, any]{
			Config: config.ConfigSpec{
				WorkflowNamespace: "ignored-legacy-namespace",
			},
		}
		meta := w.CommonWorkflowMeta(owner, *provider)

		Expect(meta.Namespace).To(Equal("provider-namespace"))
	})
})

var _ = Describe("checkResourceNamespaceAllowed", func() {
	It("fails when the resource namespace is not in the provider allowed namespaces", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{"bar"}
		provider.Name = "test"
		err := checkResourceNamespaceAllowed(types.NamespacedName{
			Namespace: "foo",
			Name:      "test-resource",
		}, *provider)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("resource test-resource in namespace foo is not allowed by provider test"))
	})

	It("succeeds when the provider allowed namespaces is empty", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{}
		err := checkResourceNamespaceAllowed(types.NamespacedName{
			Namespace: "foo",
			Name:      "test-resource",
		}, *provider)
		Expect(err).To(Not(HaveOccurred()))
	})

	It("succeeds when the resource namespace is in the provider allowed namespaces", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{"foo"}
		err := checkResourceNamespaceAllowed(types.NamespacedName{
			Namespace: "foo",
			Name:      "test-resource",
		}, *provider)
		Expect(err).To(Not(HaveOccurred()))
	})
})
