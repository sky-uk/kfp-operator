package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

type RunConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace,omitempty"`
	ProviderStatusNamespace string `json:"providerStatusNamespace,omitempty"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return rcr.ProviderNamespace == "" && rcr.ProviderStatusNamespace == ""
}

func (RunConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunScheduleConversionRemainder struct {
	ProviderNamespace       string       `json:"providerNamespace,omitempty"`
	ProviderStatusNamespace string       `json:"providerStatusNamespace,omitempty"`
	Schedule                hub.Schedule `json:"schedule,omitempty"`
}

func (rscr RunScheduleConversionRemainder) Empty() bool {
	return rscr.ProviderNamespace == "" &&
		rscr.ProviderStatusNamespace == "" &&
		rscr.Schedule.Empty()
}

func (RunScheduleConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	ProviderNamespace       string         `json:"providerNamespace,omitempty"`
	ProviderStatusNamespace string         `json:"providerStatusNamespace,omitempty"`
	Schedules               []hub.Schedule `json:"schedules,omitempty"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	for _, schedule := range rccr.Schedules {
		if !schedule.Empty() {
			return false
		}
	}

	return len(rccr.Schedules) == 0 &&
		rccr.ProviderStatusNamespace == "" &&
		rccr.ProviderNamespace == ""
}

func (RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ExperimentConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace,omitempty"`
	ProviderStatusNamespace string `json:"providerStatusNamespace,omitempty"`
}

func (ecr ExperimentConversionRemainder) Empty() bool {
	return ecr.ProviderNamespace == "" && ecr.ProviderStatusNamespace == ""
}

func (ExperimentConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type PipelineConversionRemainder struct {
	ProviderNamespace       string                `json:"providerNamespace,omitempty"`
	ProviderStatusNamespace string                `json:"providerStatusNamespace,omitempty"`
	Framework               hub.PipelineFramework `json:"framework"`
}

func (pcr PipelineConversionRemainder) Empty() bool {
	return pcr.ProviderNamespace == "" &&
		pcr.ProviderStatusNamespace == "" &&
		pcr.Framework.Name == ""
}

func (PipelineConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ProviderConversionRemainder struct {
	Image             string   `json:"image"`
	ServiceImage      string   `json:"serviceImage"`
	AllowedNamespaces []string `json:"allowedNamespaces"`
}

func (pcr ProviderConversionRemainder) Empty() bool {
	return pcr.Image == "" && pcr.ServiceImage == "" && len(pcr.AllowedNamespaces) == 0
}

func (ProviderConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}
