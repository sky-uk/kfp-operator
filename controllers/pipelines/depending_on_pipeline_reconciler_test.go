//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
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

	DescribeTable("dependentPipelineVersionIfSucceeded", func(pipeline *pipelinesv1.Pipeline, expectedVersion string, expectedSetVersion bool) {
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
