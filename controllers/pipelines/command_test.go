//go:build unit
// +build unit

package pipelines

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"math/rand"
)

var _ = Describe("eventMessage", func() {
	DescribeTable("Prints the state and the version", func(state apis.SynchronizationState) {
		version := RandomString()
		Expect(
			eventMessage(*NewSetStatus().
				WithSynchronizationState(state).
				WithVersion(version),
			),
		).To(Equal(fmt.Sprintf(`%s [version: "%s"]`, string(state), version)))
	},
		Entry("Creating", apis.Creating),
		Entry("Succeeded", apis.Succeeded),
		Entry("Updating", apis.Updating),
		Entry("Deleting", apis.Deleting),
		Entry("Deleted", apis.Deleted),
		Entry("Failed", apis.Failed),
	)

	DescribeTable("Appends a message when provided", func(state apis.SynchronizationState) {
		message := RandomString()

		Expect(
			eventMessage(*NewSetStatus().
				WithSynchronizationState(state).
				WithMessage(message),
			),
		).To(HaveSuffix(fmt.Sprintf(": %s", message)))
	},
		Entry("Creating", apis.Creating),
		Entry("Succeeded", apis.Succeeded),
		Entry("Updating", apis.Updating),
		Entry("Deleting", apis.Deleting),
		Entry("Deleted", apis.Deleted),
		Entry("Failed", apis.Failed),
	)
})

var _ = Describe("eventType", func() {
	DescribeTable("is 'Normal' for all states but 'Failed'", func(state apis.SynchronizationState) {
		Expect(
			eventType(*NewSetStatus().WithSynchronizationState(state)),
		).To(Equal(EventTypes.Normal))
	},
		Entry("Creating", apis.Creating),
		Entry("Succeeded", apis.Succeeded),
		Entry("Updating", apis.Updating),
		Entry("Deleting", apis.Deleting),
		Entry("Deleted", apis.Deleted),
	)

	When("called on 'Failed'", func() {
		It("results in 'Warning'", func() {
			Expect(
				eventType(*NewSetStatus().WithSynchronizationState(apis.Failed)),
			).To(Equal(EventTypes.Warning))
		})
	})
})

var _ = Describe("eventReason", func() {
	DescribeTable("is the expected reason for any given SynchronizationState", func(state apis.SynchronizationState, expectedReason string) {
		Expect(
			eventReason(*NewSetStatus().WithSynchronizationState(state)),
		).To(Equal(expectedReason))
	},
		Entry("Creating", apis.Creating, EventReasons.Syncing),
		Entry("Succeeded", apis.Succeeded, EventReasons.Synced),
		Entry("Updating", apis.Updating, EventReasons.Syncing),
		Entry("Deleting", apis.Deleting, EventReasons.Syncing),
		Entry("Deleted", apis.Deleted, EventReasons.Synced),
		Entry("Failed", apis.Failed, EventReasons.SyncFailed),
	)
})

var _ = Describe("alwaysSetObservedGeneration", func() {
	It("updates existing SetStatus", func() {
		commands := []Command{
			AcquireResource{},
			SetStatus{
				Status: apis.Status{SynchronizationState: apis.Succeeded},
			},
			ReleaseResource{},
		}
		resource := &pipelinesv1.Pipeline{
			Status: apis.Status{
				ObservedGeneration: -1,
			},
		}
		resource.SetGeneration(rand.Int63())

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource)

		Expect(modifiedCommands).To(Equal(
			[]Command{
				AcquireResource{},
				SetStatus{
					Status: apis.Status{
						SynchronizationState: apis.Succeeded,
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
			Status: apis.Status{
				ObservedGeneration: generation,
			},
		}
		resource.SetGeneration(generation)

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource)

		Expect(modifiedCommands).To(Equal(commands))
	})
})
