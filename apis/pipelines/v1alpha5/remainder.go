package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type RunConversionRemainder struct {
	Provider common.NamespacedName `json:"provider,omitempty"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return rcr.Provider.Empty()
}

func (RunConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunScheduleConversionRemainder struct {
	Provider common.NamespacedName `json:"provider,omitempty"`
	Schedule hub.Schedule          `json:"schedule,omitempty"`
}

func (rscr RunScheduleConversionRemainder) Empty() bool {
	return rscr.Provider.Empty() && rscr.Schedule.Empty()
}

func (RunScheduleConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	Provider  common.NamespacedName `json:"provider,omitempty"`
	Schedules []hub.Schedule        `json:"schedules,omitempty"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	for _, schedule := range rccr.Schedules {
		if schedule.Empty() {
			return false
		}
	}
	return len(rccr.Schedules) == 0 && rccr.Empty()
}

func (RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ExperimentConversionRemainder struct {
	Provider common.NamespacedName `json:"provider,omitempty"`
}

func (ecr ExperimentConversionRemainder) Empty() bool {
	return ecr.Provider.Empty()
}

func (ExperimentConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type PipelineConversionRemainder struct {
	Provider  common.NamespacedName `json:"provider,omitempty"`
	Framework hub.PipelineFramework `json:"framework"`
}

func (pcr PipelineConversionRemainder) Empty() bool {
	return pcr.Provider.Empty() && pcr.Framework.Type == ""
}

func (PipelineConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}
