//go:build unit

package workflowfactory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("runDefinitionBuilder", func() {
	newRun := func() *pipelineshub.Run {
		run := pipelineshub.RandomRun(common.RandomNamespacedName())
		run.ObjectMeta = metav1.ObjectMeta{
			Name:      "runName",
			Namespace: "runNamespace",
		}
		return run
	}

	Context("build", func() {
		builder := runDefinitionBuilder{config: config.ConfigSpec{}}

		When("the Run specifies an experiment name", func() {
			It("sets the run namespace on the experiment name", func() {
				run := newRun()
				run.Spec.ExperimentName = "myExperiment"

				runDefinition, err := builder.build(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(runDefinition.ExperimentName).To(Equal(common.NamespacedName{
					Name:      "myExperiment",
					Namespace: "runNamespace",
				}))
			})
		})

		When("the Run does not specify an experiment name", func() {
			It("leaves the experiment name empty, without a namespace", func() {
				run := newRun()
				run.Spec.ExperimentName = ""

				runDefinition, err := builder.build(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(runDefinition.ExperimentName).To(Equal(common.NamespacedName{}))
			})
		})
	})
})
