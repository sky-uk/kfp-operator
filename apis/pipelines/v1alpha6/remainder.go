package v1alpha6

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RunScheduleConversionRemainder struct {
	Schedule Schedule `json:"schedule,omitempty"`
}

func (s Schedule) Empty() bool {
	return s.StartTime == metav1.Time{} && s.EndTime == metav1.Time{}
}

func (rscr RunScheduleConversionRemainder) Empty() bool {
	return rscr.Schedule.Empty()
}

func (rscr RunScheduleConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	Schedules []Schedule `json:"schedules,omitempty"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	for _, schedule := range rccr.Schedules {
		if !schedule.Empty() {
			return false
		}
	}
	return len(rccr.Schedules) > 0
}

func (rccr RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
