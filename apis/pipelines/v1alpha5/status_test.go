//go:build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
)

var _ = Context("Status", func() {
	var _ = Describe("WithCondition", func() {
		Specify("Overrides an existing condition if the reason has changed", func() {
			status := Status{
				Conditions: []v1.Condition{
					{
						Reason: apis.RandomString(),
					},
				},
			}

			newCondition := v1.Condition{
				Reason: apis.RandomString(),
			}

			Expect(status.WithCondition(newCondition).Conditions).To(ConsistOf(newCondition))
		})

		Specify("Overrides an existing condition if the status has changed", func() {
			status := Status{
				Conditions: []v1.Condition{
					{
						Status: v1.ConditionStatus(apis.RandomString()),
					},
				},
			}

			newCondition := v1.Condition{
				Status: apis.RandomConditionStatus(),
			}

			Expect(status.WithCondition(newCondition).Conditions).To(ConsistOf(newCondition))
		})

		Specify("Overrides an existing condition if the observedGeneration has changed", func() {
			status := Status{
				Conditions: []v1.Condition{
					{
						ObservedGeneration: rand.Int63(),
					},
				},
			}

			newCondition := v1.Condition{
				ObservedGeneration: rand.Int63(),
			}

			Expect(status.WithCondition(newCondition).Conditions).To(ConsistOf(newCondition))
		})

		Specify("Keeps existing condition if neither the reason nor the status nor the observedGeneration have changed", func() {
			oldCondition := v1.Condition{
				Status:             apis.RandomConditionStatus(),
				Reason:             apis.RandomString(),
				ObservedGeneration: rand.Int63(),
				LastTransitionTime: v1.Now(),
			}

			status := Status{
				Conditions: []v1.Condition{
					oldCondition,
				},
			}

			newCondition := v1.Condition{
				Status:             oldCondition.Status,
				Reason:             oldCondition.Reason,
				ObservedGeneration: oldCondition.ObservedGeneration,
				LastTransitionTime: v1.Now(),
				Message:            apis.RandomString(),
			}

			Expect(status.WithCondition(newCondition).Conditions).To(ConsistOf(oldCondition))
		})

		Specify("Keeps other conditions unchanged", func() {
			oldConditions := apis.RandomList(func() v1.Condition {
				return v1.Condition{
					Type: apis.RandomString(),
				}
			})

			status := Status{
				Conditions: oldConditions,
			}

			newCondition := v1.Condition{
				Type: apis.RandomString(),
			}

			Expect(status.WithCondition(newCondition).Conditions).To(ContainElements(oldConditions))
		})
	})
})
