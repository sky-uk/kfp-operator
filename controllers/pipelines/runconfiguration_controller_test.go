//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var _ = Describe("RunConfiguration Controller", func() {
		v0 := RandomString()
		v1 := RandomString()
		empty := ""

		DescribeTable("newDependentPipelineVersion", func(state pipelinesv1.SynchronizationState, desiredVersion string, pipelineVersion string, expectedVersion *string) {
			rc := pipelinesv1.RunConfiguration{
				Status: pipelinesv1.RunConfigurationStatus{
					DesiredPipelineVersion: desiredVersion,
				},
			}

			pipeline := pipelinesv1.Pipeline{
				Status: pipelinesv1.Status{
					SynchronizationState: state,
					Version: pipelineVersion,
				},
			}
			Expect(newDependentPipelineVersion(&rc, &pipeline)).To(Equal(expectedVersion))
		},
		Entry(nil, pipelinesv1.Succeeded, v0, v0, nil),
		Entry(nil, pipelinesv1.Succeeded, v0, v1, &v1),
		Entry(nil, pipelinesv1.Succeeded, v0, empty, &empty),
		Entry(nil, pipelinesv1.Deleting,  v0, v0, &empty),
		Entry(nil, pipelinesv1.Deleting,  empty, v0, nil),
		Entry(nil, pipelinesv1.Deleting,  v0, v1, &empty),
		Entry(nil, pipelinesv1.Deleted,  v0, v0, &empty),
		Entry(nil, pipelinesv1.Deleted,  empty, v0, nil),
		Entry(nil, pipelinesv1.Deleted,  v0, v1, &empty),
		Entry(nil, pipelinesv1.Creating, v0, v0, nil),
		Entry(nil, pipelinesv1.Creating, v0, v1, nil),
		Entry(nil, pipelinesv1.Creating, v0, empty, nil),
		Entry(nil, pipelinesv1.Updating, v0, v0, nil),
		Entry(nil, pipelinesv1.Updating, v0, v1, nil),
		Entry(nil, pipelinesv1.Updating, v0, empty, nil),
		Entry(nil, pipelinesv1.Failed, v0, v0, nil),
		Entry(nil, pipelinesv1.Failed, v0, v1, nil),
		Entry(nil, pipelinesv1.Failed, v0, empty, nil),
	)
})
