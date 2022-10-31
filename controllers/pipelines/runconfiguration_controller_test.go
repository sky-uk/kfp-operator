//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Describe("RunConfiguration Controller", func() {
	empty := ""
	version := apis.RandomString()

	pipelineInState := func(state apis.SynchronizationState) *pipelinesv1.Pipeline {
		return &pipelinesv1.Pipeline{
			Status: pipelinesv1.Status{
				SynchronizationState: state,
				Version:              version,
			},
		}
	}

	DescribeTable("dependentPipelineVersionIfStable", func(pipeline *pipelinesv1.Pipeline, expectedVersion *string) {
		Expect(dependentPipelineVersionIfStable(pipeline)).To(Equal(expectedVersion))
	},
		Entry(nil, pipelineInState(apis.Succeeded), &version),
		Entry(nil, pipelineInState(apis.Deleted), &empty),
		Entry(nil, nil, &empty),
		Entry(nil, pipelineInState(apis.Creating), nil),
		Entry(nil, pipelineInState(apis.Updating), nil),
		Entry(nil, pipelineInState(apis.Deleting), nil),
		Entry(nil, pipelineInState(apis.Failed), nil),
	)
})
