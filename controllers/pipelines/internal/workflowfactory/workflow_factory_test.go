//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		owner := pipelineshub.RandomResource()
		namespace := RandomString()
		w := ResourceWorkflowFactory[*pipelineshub.TestResource, any]{
			Config: config.KfpControllerConfigSpec{
				WorkflowNamespace: namespace,
			},
		}
		meta := w.CommonWorkflowMeta(owner)

		Expect(meta.Namespace).To(Equal(namespace))
		Expect(meta.GetGenerateName()).To(Equal(owner.GetKind() + "-" + owner.GetName() + "-"))

		Expect(meta.Labels[workflowconstants.OwnerKindLabelKey]).To(Equal(owner.GetKind()))
		Expect(meta.Labels[workflowconstants.OwnerNameLabelKey]).To(Equal(owner.GetName()))
		Expect(meta.Labels[workflowconstants.OwnerNamespaceLabelKey]).To(Equal(owner.GetNamespace()))
	})

	It("uses config.WorkflowNamespace if set", func() {
		owner := pipelineshub.RandomResource()
		configuredNamespace := "configuredNamespace"
		w := ResourceWorkflowFactory[*pipelineshub.TestResource, any]{
			Config: config.KfpControllerConfigSpec{
				WorkflowNamespace: configuredNamespace,
			},
		}
		meta := w.CommonWorkflowMeta(owner)

		Expect(meta.Namespace).To(Equal(configuredNamespace))
	})
})

var _ = Describe("checkResourceNamespaceAllowed", func() {
	It("fails when the resource namespace is not in the provider allowed namespaces", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{"bar"}
		provider.Name = "test"
		err := checkResourceNamespaceAllowed("foo", *provider)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("resource namespace foo is not allowed by provider test"))
	})

	It("succeeds when the provider allowed namespaces is empty", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{}
		err := checkResourceNamespaceAllowed("foo", *provider)
		Expect(err).To(Not(HaveOccurred()))
	})

	It("succeeds when the resource namespace is in the provider allowed namespaces", func() {
		provider := pipelineshub.RandomProvider()
		provider.Spec.AllowedNamespaces = []string{"foo"}
		err := checkResourceNamespaceAllowed("foo", *provider)
		Expect(err).To(Not(HaveOccurred()))
	})
})
