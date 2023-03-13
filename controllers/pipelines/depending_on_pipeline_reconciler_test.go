//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Describe("DependingOnPipelineReconciler", func() {
	version := apis.RandomString()

	pipelineInState := func(state apis.SynchronizationState) *pipelinesv1.Pipeline {
		return &pipelinesv1.Pipeline{
			Status: pipelinesv1.Status{
				SynchronizationState: state,
				Version:              version,
			},
		}
	}

	DescribeTable("dependentPipelineVersionIfSucceeded", func(pipeline *pipelinesv1.Pipeline, expectedSetVersion bool, expectedVersion string) {
		setVersion, version := dependentPipelineVersionIfSucceeded(pipeline)
		Expect(setVersion).To(Equal(expectedSetVersion))
		Expect(version).To(Equal(version))
	},
		Entry(nil, pipelineInState(apis.Succeeded), true, version),
		Entry(nil, pipelineInState(apis.Deleted), true, ""),
		Entry(nil, nil, true, ""),
		Entry(nil, pipelineInState(apis.Creating), false, ""),
		Entry(nil, pipelineInState(apis.Updating), false, ""),
		Entry(nil, pipelineInState(apis.Deleting), false, ""),
		Entry(nil, pipelineInState(apis.Failed), false, ""),
	)
})
