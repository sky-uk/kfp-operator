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
				{Type: "a"},
				RandomCronTrigger(),
				{Type: "b"},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(2))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeNotSupported))
			Expect(errors[1].Type).To(Equal(field.ErrorTypeNotSupported))
		})

		Specify("requires cron expression for schedule triggers", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{Type: TriggerTypes.Schedule},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(1))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeRequired))
			Expect(errors[0].Field).To(Equal("spec.triggers[0].cronExpression"))
			Expect(errors[0].Detail).To(Equal("required for trigger type schedule"))
		})

		Specify("rejects cron expression for non-schedule triggers", func() {
			rc := RandomRunConfiguration()
			rc.Spec.Triggers = []Trigger{
				{Type: TriggerTypes.OnChange},
				{Type: TriggerTypes.OnChange, CronExpression: "1 2 3 4 5"},
			}

			errors := rc.errorsInTriggers()
			Expect(errors).To(HaveLen(1))
			Expect(errors[0].Type).To(Equal(field.ErrorTypeForbidden))
			Expect(errors[0].Field).To(Equal("spec.triggers[1].cronExpression"))
			Expect(errors[0].Detail).To(Equal("not allowed for trigger type onChange"))
		})
	})
})
