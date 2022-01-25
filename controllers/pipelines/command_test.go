//go:build unit
// +build unit

package pipelines

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var _ = Describe("eventMessage", func() {
	DescribeTable("Prints the state and the version", func(state pipelinesv1.SynchronizationState) {
		version := RandomString()
		Expect(
			eventMessage(*NewSetStatus().
				WithSynchronizationState(state).
				WithVersion(version),
			),
		).To(Equal(fmt.Sprintf(`%s [version: "%s"]`, string(state), version)))
	},
		Entry("Creating", pipelinesv1.Creating),
		Entry("Succeeded", pipelinesv1.Succeeded),
		Entry("Updating", pipelinesv1.Updating),
		Entry("Deleting", pipelinesv1.Deleting),
		Entry("Deleted", pipelinesv1.Deleted),
		Entry("Failed", pipelinesv1.Failed),
	)

	DescribeTable("Appends a message when provided", func(state pipelinesv1.SynchronizationState) {
		message := RandomString()

		Expect(
			eventMessage(*NewSetStatus().
				WithSynchronizationState(state).
				WithMessage(message),
			),
		).To(HaveSuffix(fmt.Sprintf(": %s", message)))
	},
		Entry("Creating", pipelinesv1.Creating),
		Entry("Succeeded", pipelinesv1.Succeeded),
		Entry("Updating", pipelinesv1.Updating),
		Entry("Deleting", pipelinesv1.Deleting),
		Entry("Deleted", pipelinesv1.Deleted),
		Entry("Failed", pipelinesv1.Failed),
	)
})

var _ = Describe("eventType", func() {
	DescribeTable("is 'Normal' for all states but 'Failed'", func(state pipelinesv1.SynchronizationState) {
		Expect(
			eventType(*NewSetStatus().WithSynchronizationState(state)),
		).To(Equal(EventTypes.Normal))
	},
		Entry("Creating", pipelinesv1.Creating),
		Entry("Succeeded", pipelinesv1.Succeeded),
		Entry("Updating", pipelinesv1.Updating),
		Entry("Deleting", pipelinesv1.Deleting),
		Entry("Deleted", pipelinesv1.Deleted),
	)

	When("called on 'Failed'", func() {
		It("results in 'Warning'", func() {
			Expect(
				eventType(*NewSetStatus().WithSynchronizationState(pipelinesv1.Failed)),
			).To(Equal(EventTypes.Warning))
		})
	})
})

var _ = Describe("eventReason", func() {
	DescribeTable("is the expected reason for any given SynchronizationState", func(state pipelinesv1.SynchronizationState, expectedReason string) {
		Expect(
			eventReason(*NewSetStatus().WithSynchronizationState(state)),
		).To(Equal(expectedReason))
	},
		Entry("Creating", pipelinesv1.Creating, EventReasons.Syncing),
		Entry("Succeeded", pipelinesv1.Succeeded, EventReasons.Synced),
		Entry("Updating", pipelinesv1.Updating, EventReasons.Syncing),
		Entry("Deleting", pipelinesv1.Deleting, EventReasons.Syncing),
		Entry("Deleted", pipelinesv1.Deleted, EventReasons.Synced),
		Entry("Failed", pipelinesv1.Failed, EventReasons.SyncFailed),
	)
})
