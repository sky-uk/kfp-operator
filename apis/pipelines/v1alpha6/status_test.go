//go:build unit

package v1alpha6

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
)

var _ = Context("Conditions", func() {
	var _ = Describe("MergeIntoConditions", func() {
		Specify("Overrides an existing condition if the reason has changed", func() {
			conditions := Conditions{
				{
					Reason: apis.RandomString(),
				},
			}

			newCondition := metav1.Condition{
				Reason: apis.RandomString(),
			}

			Expect(conditions.MergeIntoConditions(newCondition)).To(ConsistOf(newCondition))
		})

		Specify("Overrides an existing condition if the status has changed", func() {
			conditions := Conditions{
				{
					Status: metav1.ConditionStatus(apis.RandomString()),
				},
			}

			newCondition := metav1.Condition{
				Status: apis.RandomConditionStatus(),
			}

			Expect(conditions.MergeIntoConditions(newCondition)).To(ConsistOf(newCondition))
		})

		Specify("Overrides an existing condition if the observedGeneration has changed", func() {
			conditions := Conditions{
				{
					ObservedGeneration: rand.Int63(),
				},
			}

			newCondition := metav1.Condition{
				ObservedGeneration: rand.Int63(),
			}

			Expect(conditions.MergeIntoConditions(newCondition)).To(ConsistOf(newCondition))
		})

		Specify("Keeps existing condition if neither the reason nor the status nor the observedGeneration have changed", func() {
			oldCondition := metav1.Condition{
				Status:             apis.RandomConditionStatus(),
				Reason:             apis.RandomString(),
				ObservedGeneration: rand.Int63(),
				LastTransitionTime: metav1.Now(),
			}

			conditions := Conditions{
				oldCondition,
			}

			newCondition := metav1.Condition{
				Status:             oldCondition.Status,
				Reason:             oldCondition.Reason,
				ObservedGeneration: oldCondition.ObservedGeneration,
				LastTransitionTime: metav1.Now(),
				Message:            apis.RandomString(),
			}

			Expect(conditions.MergeIntoConditions(newCondition)).To(ConsistOf(oldCondition))
		})

		Specify("Keeps other conditions unchanged", func() {
			oldConditions := Conditions(apis.RandomList(func() metav1.Condition {
				return metav1.Condition{
					Type: apis.RandomString(),
				}
			}))

			newCondition := metav1.Condition{
				Type: apis.RandomString(),
			}

			Expect(oldConditions.MergeIntoConditions(newCondition)).To(ContainElements(oldConditions))
		})
	})
	var _ = Describe("ConditionStatusForSynchronizationState", func() {
		DescribeTable("Converts SynchronizationState to ConditionStatus", func(state apis.SynchronizationState, status metav1.ConditionStatus) {
			Expect(ConditionStatusForSynchronizationState(state)).To(Equal(status))
		},
			Entry("", apis.Succeeded, metav1.ConditionTrue),
			Entry("", apis.Deleted, metav1.ConditionTrue),
			Entry("", apis.Failed, metav1.ConditionFalse),
			Entry("", apis.Updating, metav1.ConditionUnknown),
			Entry("", apis.SynchronizationState(""), metav1.ConditionUnknown),
			Entry("", apis.Deleting, metav1.ConditionUnknown))
	})
})
