package v1alpha6

type RunScheduleConversionRemainder struct {
	Schedule Schedule `json:"schedule,omitempty"`
}

func (s Schedule) empty() bool {
	return s.StartTime == nil && s.EndTime == nil
}

func (rscr RunScheduleConversionRemainder) Empty() bool {
	return rscr.Schedule.empty()
}

func (rscr RunScheduleConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	Schedules []Schedule `json:"schedules,omitempty"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	for _, schedule := range rccr.Schedules {
		if schedule.empty() {
			return false
		}
	}
	return len(rccr.Schedules) == 0
}

func (rccr RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
