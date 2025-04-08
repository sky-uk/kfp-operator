//go:build unit

package pipelines

import (
	"context"
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("eventMessage", func() {
	DescribeTable("Prints the state and the version", func(state apis.SynchronizationState) {
		version := apis.RandomString()
		Expect(
			eventMessage(*NewSetStatus().
				WithSyncStateCondition(state, metav1.Time{}, "").
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
		message := apis.RandomString()

		Expect(
			eventMessage(*NewSetStatus().
				WithSyncStateCondition(state, metav1.Time{}, message),
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
			eventType(*NewSetStatus().WithSyncStateCondition(state, metav1.Time{}, "")),
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
				eventType(*NewSetStatus().WithSyncStateCondition(apis.Failed, metav1.Time{}, "")),
			).To(Equal(EventTypes.Warning))
		})
	})
})

var _ = Describe("eventReason", func() {
	DescribeTable("is the expected reason for any given SynchronizationState", func(state apis.SynchronizationState, expectedReason string) {
		Expect(
			eventReason(*NewSetStatus().WithSyncStateCondition(state, metav1.Time{}, "")),
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
		transitionTime := metav1.Now()
		status := SetStatus{
			Status: pipelineshub.Status{},
		}
		status.WithSyncStateCondition(apis.Succeeded, transitionTime, "")

		commands := []Command{
			AcquireResource{},
			status,
			ReleaseResource{},
		}
		resource := &pipelineshub.Pipeline{
			Status: pipelineshub.Status{
				ObservedGeneration: -1,
			},
		}
		resource.SetGeneration(rand.Int63())
		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource, transitionTime)

		expectedSetStatus := SetStatus{
			Status: pipelineshub.Status{
				ObservedGeneration: resource.Generation,
			},
		}
		expectedSetStatus.WithSyncStateCondition(apis.Succeeded, transitionTime, "")

		Expect(modifiedCommands).To(ContainElements(
			[]Command{
				AcquireResource{},
				expectedSetStatus,
				ReleaseResource{},
			}))
	})

	It("appends SetStatus when it doesn't exist", func() {
		commands := []Command{
			AcquireResource{},
			ReleaseResource{},
		}
		resource := &pipelineshub.Pipeline{
			Status: pipelineshub.RandomStatus(common.RandomNamespacedName()),
		}
		resource.SetGeneration(rand.Int63())
		resource.Status.ObservedGeneration = -1

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource, metav1.Time{})

		expectedResource := resource.Status.DeepCopy()
		expectedResource.ObservedGeneration = resource.GetGeneration()

		expectedSetStatus := SetStatus{
			Status: *expectedResource,
		}

		Expect(modifiedCommands).To(ContainElements(
			[]Command{
				AcquireResource{},
				ReleaseResource{},
				expectedSetStatus,
			}))
	})

	It("leaves commands unchanged when the generation hasn't changed", func() {
		commands := []Command{
			AcquireResource{},
			ReleaseResource{},
		}
		generation := rand.Int63()
		resource := &pipelineshub.Pipeline{
			Status: pipelineshub.Status{
				ObservedGeneration: generation,
			},
		}
		resource.SetGeneration(generation)

		modifiedCommands := alwaysSetObservedGeneration(context.Background(), commands, resource, metav1.Time{})

		Expect(modifiedCommands).To(Equal(commands))
	})
})
