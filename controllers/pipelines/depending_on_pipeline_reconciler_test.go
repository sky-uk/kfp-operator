//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Describe("DependingOnPipelineReconciler", func() {
	version := apis.RandomString()

	pipelineInState := func(state apis.SynchronizationState) *pipelineshub.Pipeline {
		return &pipelineshub.Pipeline{
			Status: pipelineshub.Status{
				Version: version,
				Conditions: apis.Conditions{
					{
						Type:   apis.ConditionTypes.SynchronizationSucceeded,
						Status: apis.ConditionStatusForSynchronizationState(state),
						Reason: string(state),
					},
				},
			},
		}
	}

	DescribeTable("dependentPipelineVersionIfSucceeded", func(pipeline *pipelineshub.Pipeline, expectedVersion string, expectedSetVersion bool) {
		version, setVersion := dependentPipelineVersionIfSucceeded(pipeline)
		Expect(version).To(Equal(version))
		Expect(setVersion).To(Equal(expectedSetVersion))
	},
		Entry(nil, pipelineInState(apis.Succeeded), version, true),
		Entry(nil, pipelineInState(apis.Deleted), "", true),
		Entry(nil, nil, "", true),
		Entry(nil, pipelineInState(apis.Creating), "", false),
		Entry(nil, pipelineInState(apis.Updating), "", false),
		Entry(nil, pipelineInState(apis.Deleting), "", false),
		Entry(nil, pipelineInState(apis.Failed), "", false),
	)
})
