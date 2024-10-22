package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func convertRuntimeParametersToHub(rtp []RuntimeParameter) (hubRtp []hub.RuntimeParameter) {
	for _, namedValue := range rtp {
		hubRtp = append(hubRtp, hub.RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}

	return hubRtp
}

func convertArtifactLocatorToHub(artifactLocator ArtifactLocator) hub.ArtifactLocator {
	return hub.ArtifactLocator{
		Component: artifactLocator.Component,
		Artifact:  artifactLocator.Artifact,
		Index:     artifactLocator.Index,
	}
}

func convertArtifactPathToHub(artifactPath ArtifactPath) hub.ArtifactPath {
	return hub.ArtifactPath{
		Locator: convertArtifactLocatorToHub(artifactPath.Locator),
		Filter:  artifactPath.Filter,
	}
}

func convertArtifactsToHub(artifacts []OutputArtifact) (hubArtifacts []hub.OutputArtifact) {
	for _, artifact := range artifacts {
		hubArtifacts = append(hubArtifacts, hub.OutputArtifact{
			Name: artifact.Name,
			Path: convertArtifactPathToHub(artifact.Path),
		})
	}

	return hubArtifacts
}

func convertScheduleTriggersToHub(schedules []string) (hubSchedules []hub.Schedule) {
	for _, schedule := range schedules {
		hubSchedules = append(hubSchedules, hub.Schedule{
			CronExpression: schedule,
		})
	}
	return hubSchedules
}

func convertOnChangeTriggersToHub(oct []OnChangeType) (hubOct []hub.OnChangeType) {
	for _, onChange := range oct {
		hubOct = append(hubOct, hub.OnChangeType(onChange))
	}
	return hubOct
}

func convertTriggersToHub(triggers Triggers) hub.Triggers {
	return hub.Triggers{
		Schedules:         convertScheduleTriggersToHub(triggers.Schedules),
		OnChange:          convertOnChangeTriggersToHub(triggers.OnChange),
		RunConfigurations: triggers.RunConfigurations,
	}
}

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersToHub(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsToHub(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersToHub(src.Spec.Triggers)

	dst.Status = hub.RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}

	return nil
}

func convertRuntimeParametersFromHub(hubRtp []hub.RuntimeParameter) (rtp []RuntimeParameter) {
	for _, namedValue := range hubRtp {
		rtp = append(rtp, RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}

	return rtp
}

func convertArtifactLocatorFromHub(hubArtifactLocator hub.ArtifactLocator) ArtifactLocator {
	return ArtifactLocator{
		Component: hubArtifactLocator.Component,
		Artifact:  hubArtifactLocator.Artifact,
		Index:     hubArtifactLocator.Index,
	}
}

func convertArtifactPathFromHub(hubArtifactPath hub.ArtifactPath) ArtifactPath {
	return ArtifactPath{
		Locator: convertArtifactLocatorFromHub(hubArtifactPath.Locator),
		Filter:  hubArtifactPath.Filter,
	}
}

func convertArtifactsFromHub(hubArtifacts []hub.OutputArtifact) (artifacts []OutputArtifact) {
	for _, artifact := range hubArtifacts {
		artifacts = append(artifacts, OutputArtifact{
			Name: artifact.Name,
			Path: convertArtifactPathFromHub(artifact.Path),
		})
	}

	return artifacts
}

func convertScheduleTriggersFromHub(hubSchedules []hub.Schedule) (schedules []string) {
	for _, schedule := range hubSchedules {
		schedules = append(schedules, schedule.CronExpression)
	}
	return schedules
}

func convertOnChangeTriggersFromHub(hubOct []hub.OnChangeType) (oct []OnChangeType) {
	for _, onChange := range hubOct {
		oct = append(oct, OnChangeType(onChange))
	}
	return oct
}

func convertTriggersFromHub(hubTriggers hub.Triggers) Triggers {
	return Triggers{
		Schedules:         convertScheduleTriggersFromHub(hubTriggers.Schedules),
		OnChange:          convertOnChangeTriggersFromHub(hubTriggers.OnChange),
		RunConfigurations: hubTriggers.RunConfigurations,
	}
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersFromHub(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsFromHub(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersFromHub(src.Spec.Triggers)
	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}
	return nil
}
