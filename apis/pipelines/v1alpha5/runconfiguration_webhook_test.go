//go:build unit
// +build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Context("RunConfiguration Webhook", func() {
	var _ = Describe("errorsInTriggers", func() {

		Specify("returns no errors for well formatted triggers", func() {
			rc := RandomRunConfiguration()

			Expect(rc.errorsInTriggers()).To(BeEmpty())
		})

		Specify("validates every trigger", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{Schedule: &ScheduleTrigger{}, OnChange: &OnChangeTrigger{}},
				{},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(3))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeRequired))
			Expect(errors[1].Type).To(Equal(field.ErrorTypeTooMany))
			Expect(errors[2].Type).To(Equal(field.ErrorTypeRequired))
		})

		Specify("requires cron expression for schedule triggers", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{Schedule: &ScheduleTrigger{}},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(1))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeRequired))
			Expect(errors[0].Field).To(Equal("spec.triggers[0].schedule.cronExpression"))
			Expect(errors[0].Detail).To(Equal("required for trigger type schedule"))
		})

		Specify("rejects more than one trigger", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{Schedule: &ScheduleTrigger{CronExpression: "1 2 3 4 5"}, OnChange: &OnChangeTrigger{}},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(1))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeTooMany))
			Expect(errors[0].Field).To(Equal("spec.triggers[0]"))
			Expect(errors[0].Detail).To(Equal("must have at most 1 items"))
		})

		Specify("rejects no trigger", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(1))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeRequired))
			Expect(errors[0].Field).To(Equal("spec.triggers[0]"))
			Expect(errors[0].Detail).To(Equal("a trigger must be set"))
		})
	})
})
