//go:build unit

package v1alpha6

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"k8s.io/apimachinery/pkg/api/errors"
)

var _ = Context("RunConfiguration Webhook", func() {
	Specify("Duplicate schedule triggers fail the validation", func() {
		runConfiguration := RunConfiguration{
			Spec: RunConfigurationSpec{
				Triggers: Triggers{
					Schedules: []Schedule{
						{CronExpression: "a"},
						{CronExpression: "a"},
					},
				},
			},
		}

		Expect(errors.IsInvalid(runConfiguration.validate())).To(BeTrue())
	})

	Specify("Duplicate onChange triggers fail the validation", func() {
		runConfiguration := RunConfiguration{
			Spec: RunConfigurationSpec{
				Triggers: Triggers{
					OnChange: []OnChangeType{
						OnChangeTypes.Pipeline,
						OnChangeTypes.Pipeline,
					},
				},
			},
		}

		Expect(errors.IsInvalid(runConfiguration.validate())).To(BeTrue())
	})

	Specify("Specifying Value and ValueFrom in the runtime parameters fails the validation", func() {
		runConfiguration := RunConfiguration{
			Spec: RunConfigurationSpec{
				Run: RunSpec{
					RuntimeParameters: []RuntimeParameter{
						{
							Name:  apis.RandomString(),
							Value: apis.RandomString(),
						},
						{
							Name:      apis.RandomString(),
							ValueFrom: &ValueFrom{},
						},
						{
							Name:      apis.RandomString(),
							Value:     apis.RandomString(),
							ValueFrom: &ValueFrom{},
						},
					},
				},
			},
		}

		Expect(errors.IsInvalid(runConfiguration.validate())).To(BeTrue())
	})

	Specify("A valid spec passes the validation", func() {
		runConfiguration := RunConfiguration{
			Spec: RunConfigurationSpec{
				Run: RunSpec{
					RuntimeParameters: []RuntimeParameter{
						{
							Name:  apis.RandomString(),
							Value: apis.RandomString(),
						},
						{
							Name:      apis.RandomString(),
							ValueFrom: &ValueFrom{},
						},
					},
				},
				Triggers: Triggers{
					Schedules: []Schedule{
						RandomSchedule(),
					},
					OnChange: []OnChangeType{
						OnChangeTypes.Pipeline,
					},
				},
			},
		}

		Expect(runConfiguration.validate()).To(Succeed())
	})
})
