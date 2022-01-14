package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

var _ = Describe("WorkflowFactory.Annotations", func() {
	When("no debug annotation is present on the ObjectMeta", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation can't be deserialized", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": "broken:",
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation is valid and ObjectMeta overrides the default config", func() {
		It("uses the overrides", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: false,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":true}`,
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})

	When("the annotation is valid and the ObjectMeta does not override the default config", func() {
		It("uses the configured defaults", func() {
			workflowFactory := WorkflowFactory{
				Config: configv1.Configuration{
					Debug: pipelinesv1.DebugOptions{
						KeepWorkflows: true,
					},
				},
			}

			objectMeta := metav1.ObjectMeta{Annotations: map[string]string{
				"pipelines.kubeflow.org/debug": `{"keepWorkflows":false}`,
			}}

			debugAnnotations := workflowFactory.Annotations(context.Background(), objectMeta)["pipelines.kubeflow.org/debug"]
			Expect(debugAnnotations).To(MatchJSON(`{"keepWorkflows":true}`))
		})
	})
})
