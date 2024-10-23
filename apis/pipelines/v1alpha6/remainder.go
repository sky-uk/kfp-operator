package v1alpha6

type RunScheduleConversionRemainder struct {
	Schedule Schedule `json:"schedule,omitempty"`
}

// TODO: check this can be private. Really shouldn't expose this function
// because it means if it is empty in the remainder sense.
func (s Schedule) Empty() bool {
	return s.StartTime == nil && s.EndTime == nil
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
		if schedule.Empty() {
			return false
		}
	}
	return len(rccr.Schedules) == 0
}

func (rccr RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
