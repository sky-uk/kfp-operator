package v1alpha6

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

type RunConversionRemainder struct {
	ProviderNamespace string `json:"providerNamespace"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return rcr.ProviderNamespace == ""
}

func (rcr RunConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunScheduleConversionRemainder struct {
	ProviderNamespace string `json:"providerNamespace"`
}

func (rsr RunScheduleConversionRemainder) Empty() bool {
	return rsr.ProviderNamespace == ""
}

func (rsr RunScheduleConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	ProviderNamespace string `json:"providerNamespace"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	return rccr.ProviderNamespace == ""
}

func (rccr RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type PipelineConversionRemainder struct {
	ProviderNamespace string                `json:"providerNamespace"`
	Framework         hub.PipelineFramework `json:"framework"`
}

func (pcr PipelineConversionRemainder) Empty() bool {
	return pcr.ProviderNamespace == "" && pcr.Framework.Type == ""
}

func (pcr PipelineConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ExperimentConversionRemainder struct {
	ProviderNamespace string `json:"providerNamespace"`
}

func (er ExperimentConversionRemainder) Empty() bool {
	return er.ProviderNamespace == ""
}

func (er ExperimentConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}
