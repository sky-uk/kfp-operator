package v1alpha6

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

type RunConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace"`
	ProviderStatusNamespace string `json:"providerStatusNamespace"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return rcr.ProviderNamespace == "" && rcr.ProviderStatusNamespace == ""
}

func (RunConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunScheduleConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace"`
	ProviderStatusNamespace string `json:"providerStatusNamespace"`
}

func (rsr RunScheduleConversionRemainder) Empty() bool {
	return rsr.ProviderNamespace == "" && rsr.ProviderStatusNamespace == ""
}

func (RunScheduleConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type RunConfigurationConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace"`
	ProviderStatusNamespace string `json:"providerStatusNamespace"`
}

func (rccr RunConfigurationConversionRemainder) Empty() bool {
	return rccr.ProviderNamespace == "" && rccr.ProviderStatusNamespace == ""
}

func (RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type PipelineConversionRemainder struct {
	ProviderNamespace       string                `json:"providerNamespace"`
	ProviderStatusNamespace string                `json:"providerStatusNamespace"`
	Framework               hub.PipelineFramework `json:"framework"`
}

func (pcr PipelineConversionRemainder) Empty() bool {
	return pcr.ProviderNamespace == "" && pcr.Framework.Name == "" && pcr.ProviderStatusNamespace == ""
}

func (PipelineConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ExperimentConversionRemainder struct {
	ProviderNamespace       string `json:"providerNamespace"`
	ProviderStatusNamespace string `json:"providerStatusNamespace"`
}

func (er ExperimentConversionRemainder) Empty() bool {
	return er.ProviderNamespace == "" && er.ProviderStatusNamespace == ""
}

func (ExperimentConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}

type ProviderConversionRemainder struct {
	Image             string   `json:"image"`
	AllowedNamespaces []string `json:"allowedNamespaces"`
}

func (pcr ProviderConversionRemainder) Empty() bool {
	return pcr.Image == "" && len(pcr.AllowedNamespaces) == 0
}

func (ProviderConversionRemainder) ConversionAnnotation() string {
	return hub.GroupVersion.Version + "." + hub.GroupVersion.Group + "/conversions.remainder"
}
