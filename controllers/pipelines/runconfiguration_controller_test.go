//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var _ = Describe("RunConfiguration Controller", func() {
	empty := ""
	version := RandomString()

	pipelineInState := func(state pipelinesv1.SynchronizationState) *pipelinesv1.Pipeline {
		return &pipelinesv1.Pipeline{
			Status: pipelinesv1.Status{
				SynchronizationState: state,
				Version:              version,
			},
		}
	}

	DescribeTable("dependentPipelineVersion", func(pipeline *pipelinesv1.Pipeline, expectedVersion *string) {
		Expect(dependentPipelineVersion(pipeline)).To(Equal(expectedVersion))
	},
		Entry(nil, pipelineInState(pipelinesv1.Succeeded), &version),
		Entry(nil, pipelineInState(pipelinesv1.Deleted), &empty),
		Entry(nil, nil, &empty),
		Entry(nil, pipelineInState(pipelinesv1.Creating), nil),
		Entry(nil, pipelineInState(pipelinesv1.Updating), nil),
		Entry(nil, pipelineInState(pipelinesv1.Deleting), nil),
		Entry(nil, pipelineInState(pipelinesv1.Failed), nil),
	)
})
