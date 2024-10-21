//go:build functional

package run_completion_event_trigger

import (
	"github.com/sky-uk/kfp-operator/argo/common"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConvertersUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Converters Unit Test Suite")
}

var _ = Context("ProtoRunCompletionToCommon", func() {
	When("given a proto run completion event", func() {
		It("returns the run completion event in the common struct", func() {
			protoRunCompletionEvent := RunCompletionEvent{
				PipelineName:         "namespace/some-pipeline",
				Provider:             "some-provider",
				RunConfigurationName: "namespace/some-run-configuration-name",
				RunId:                "some-run-id",
				RunName:              "namespace/some-run-name",
				ServingModelArtifacts: []*ServingModelArtifact{
					{
						Location: "some-location",
						Name:     "some-name",
					},
				},
				Status: Status_SUCCEEDED,
			}

			expectedCommonRunCompletionEvent := common.RunCompletionEvent{
				Status: common.RunCompletionStatuses.Succeeded,
				PipelineName: common.NamespacedName{
					Namespace: "namespace",
					Name:      "some-pipeline",
				},
				RunConfigurationName: &common.NamespacedName{
					Namespace: "namespace",
					Name:      "some-run-configuration-name",
				},
				RunName: &common.NamespacedName{
					Namespace: "namespace",
					Name:      "some-run-name",
				},
				RunId: "some-run-id",
				ServingModelArtifacts: []common.Artifact{
					{
						Location: "some-location",
						Name:     "some-name",
					},
				},
				Artifacts: nil,
				Provider:  "some-provider",
			}

			Expect(ProtoRunCompletionToCommon(&protoRunCompletionEvent)).To(Equal(expectedCommonRunCompletionEvent))
		})
	})
})

var _ = Context("ArtifactsConverter", func() {

	When("given a list of proto `ServingModelArtifacts`", func() {
		It("returns a list of artifacts in the common struct", func() {
			servingModelArtifacts := []*ServingModelArtifact{
				{
					Location: "some-location",
					Name:     "some-name",
				},
				{
					Location: "some-location-2",
					Name:     "some-name-2",
				},
			}

			Expect(ArtifactsConverter(servingModelArtifacts)).To(Equal(
				[]common.Artifact{
					{
						Location: "some-location",
						Name:     "some-name",
					},
					{
						Location: "some-location-2",
						Name:     "some-name-2",
					},
				}))
		})
	})
})

var _ = Context("StatusConverter", func() {

	When("given a `SUCCEEDED` proto Status", func() {
		It("returns a common RunCompletionStatus of succeeded", func() {
			Expect(StatusConverter(Status_SUCCEEDED)).To(Equal(common.RunCompletionStatuses.Succeeded))
		})
	})

	When("given a `FAILED` proto Status", func() {
		It("returns a common RunCompletionStatus of failed", func() {
			Expect(StatusConverter(Status_FAILED)).To(Equal(common.RunCompletionStatuses.Failed))
		})
	})

})
