package v1alpha6

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RunScheduleConversionRemainder struct {
	Schedule Schedule `json:"schedule,omitempty"`
}

func (rscr RunScheduleConversionRemainder) Empty() bool {
	return rscr.Schedule.StartTime == metav1.Time{} && rscr.Schedule.EndTime == metav1.Time{}
}

func (rscr RunScheduleConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
