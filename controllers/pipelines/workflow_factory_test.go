//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
)

var _ = Describe("CommonWorkflowMeta", func() {
	It("creates metadata", func() {
		resource := RandomNamespacedName()
		operation := RandomString()
		ownerKind := RandomString()

		meta := CommonWorkflowMeta(resource, operation, ownerKind)

		Expect(meta.Namespace).To(Equal(resource.Namespace))
		Expect(meta.GetGenerateName()).To(Equal(operation + "-" + ownerKind + "-"))

		Expect(meta.Labels[WorkflowConstants.OwnerKindLabelKey]).To(Equal(ownerKind))
		Expect(meta.Labels[WorkflowConstants.OwnerNameLabelKey]).To(Equal(resource.Name))
		Expect(meta.Labels[WorkflowConstants.OperationLabelKey]).To(Equal(operation))
	})
})

var _ = Describe("KfpExtCommandBuilder", func() {
	When("Only a command is given", func() {
		It("prints boilerplate and the command", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			kfpScript, err := workflowFactory.KfpExt("do sth").Build()

			Expect(err).NotTo(HaveOccurred())
			Expect(kfpScript).To(Equal("kfp-ext --endpoint www.example.com --output json do sth"))
		})
	})

	When("A command, a parameter, an optional parameter and an argument  are given", func() {
		It("prints boilerplate, the command, the parameters and the argument", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			kfpScript, err := workflowFactory.KfpExt("do sth").
				Param("--parameter", "value").
				OptParam("--opt-parameter", "optional value").
				Arg("anArgument").
				Build()

			Expect(err).NotTo(HaveOccurred())
			Expect(kfpScript).To(Equal(`kfp-ext --endpoint www.example.com --output json do sth --parameter 'value' --opt-parameter 'optional value' 'anArgument'`))
		})
	})

	When("An empty parameter is given", func() {
		It("errors", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			_, err := workflowFactory.KfpExt("do sth").
				Param("--param", "").
				Build()

			Expect(err).To(HaveOccurred())
		})
	})

	When("A an empty argument is given", func() {
		It("errors", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			_, err := workflowFactory.KfpExt("do sth").
				Arg("").
				Build()

			Expect(err).To(HaveOccurred())
		})
	})

	When("An empty optional parameter is given", func() {
		It("omits the parameter", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			kfpScript, err := workflowFactory.KfpExt("do sth").
				OptParam("--param", "").
				Build()

			Expect(err).NotTo(HaveOccurred())
			Expect(kfpScript).To(Equal(`kfp-ext --endpoint www.example.com --output json do sth`))
		})
	})

	When("Single quotes are given", func() {
		It("escapes the quotes", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					KfpEndpoint: "www.example.com",
				},
			}

			kfpScript, err := workflowFactory.KfpExt("do sth").
				Param("--param", "this is 'a parameter'").
				OptParam("--opt-param", "this is 'an optional parameter'").
				Arg("this is 'an argument'").
				Build()

			Expect(err).NotTo(HaveOccurred())
			Expect(kfpScript).To(Equal("kfp-ext --endpoint www.example.com --output json do sth --param 'this is \\'a parameter\\'' --opt-param 'this is \\'an optional parameter\\'' 'this is \\'an argument\\''"))
		})
	})
})
