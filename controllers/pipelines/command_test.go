//go:build unit
// +build unit

package pipelines

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"math/rand"
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

var _ = Describe("alwaysSetObservedGeneration", func() {
	It("updates existing SetStatus", func() {
		commands := []Command{
			AcquireResource{},
			SetStatus{
				Status: pipelinesv1.Status{SynchronizationState: pipelinesv1.Succeeded},
			},
			ReleaseResource{},
		}
		resource := &pipelinesv1.Pipeline{
			Status: pipelinesv1.Status{
				ObservedGeneration: -1,
			},
		}
		resource.SetGeneration(rand.Int63())

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource)

		Expect(modifiedCommands).To(Equal(
			[]Command{
				AcquireResource{},
				SetStatus{
					Status: pipelinesv1.Status{
						SynchronizationState: pipelinesv1.Succeeded,
						ObservedGeneration:   resource.Generation,
					},
				},
				ReleaseResource{},
			}))
	})

	It("appends SetStatus when it doesn't exist", func() {
		commands := []Command{
			AcquireResource{},
			ReleaseResource{},
		}
		resource := &pipelinesv1.Pipeline{
			Status: RandomStatus(),
		}
		resource.SetGeneration(rand.Int63())
		resource.Status.ObservedGeneration = -1

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource)

		expectedResource := resource.Status.DeepCopy()
		expectedResource.ObservedGeneration = resource.GetGeneration()

		Expect(modifiedCommands).To(Equal(
			[]Command{
				AcquireResource{},
				ReleaseResource{},
				SetStatus{
					Status: *expectedResource,
				},
			}))
	})

	It("leaves commands unchanged when the generation hasn't changed", func() {
		commands := []Command{
			AcquireResource{},
			ReleaseResource{},
		}
		generation := rand.Int63()
		resource := &pipelinesv1.Pipeline{
			Status: pipelinesv1.Status{
				ObservedGeneration: generation,
			},
		}
		resource.SetGeneration(generation)

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource)

		Expect(modifiedCommands).To(Equal(commands))
	})
})
