package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertRuntimeParametersTo(rtp []RuntimeParameter) []hub.RuntimeParameter {
	var hubRtp []hub.RuntimeParameter
	for _, namedValue := range rtp {
		hubRtp = append(hubRtp, hub.RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}
	return hubRtp
}

func convertArtifactLocatorTo(sl ArtifactLocator) hub.ArtifactLocator {
	return hub.ArtifactLocator{
		Component: sl.Component,
		Artifact:  sl.Artifact,
		Index:     sl.Index,
	}
}

func convertArtifactPathTo(ap ArtifactPath) hub.ArtifactPath {
	return hub.ArtifactPath{
		Locator: convertArtifactLocatorTo(ap.Locator),
		Filter:  ap.Filter,
	}
}

func convertArtifactsTo(oa []OutputArtifact) []hub.OutputArtifact {
	var hubOa []hub.OutputArtifact
	for _, artifact := range oa {
		hubOa = append(hubOa, hub.OutputArtifact{
			Name: artifact.Name,
			Path: convertArtifactPathTo(artifact.Path),
		})
	}
	return hubOa
}

func convertScheduleTo(
	schedule string,
	remainder hub.Schedule,
) (hubSchedule hub.Schedule) {
	return hub.Schedule{
		CronExpression: schedule,
		StartTime:      remainder.StartTime,
		EndTime:        remainder.EndTime,
	}
}

func convertTriggersTo(
	triggers Triggers,
	remainder []hub.Schedule,
) hub.Triggers {
	convertOnChangesTo := func(oct []OnChangeType) []hub.OnChangeType {
		var hubOct []hub.OnChangeType
		for _, onChange := range oct {
			hubOct = append(hubOct, hub.OnChangeType(onChange))
		}
		return hubOct
	}
	convertSchedulesTo := func(
		schedules []string,
		remainder []hub.Schedule,
	) []hub.Schedule {
		// Make a map of the hub CronExpression -> { StartTime, EndTime }.
		// This could potentially be lossy because if two schedules share
		// the same CronExpression, then one of them will be overwritten.
		var hubSchedules []hub.Schedule
		remainderMap := make(
			map[string]struct {
				StartTime metav1.Time
				EndTime   metav1.Time
			},
		)

		for _, schedule := range remainder {
			remainderMap[schedule.CronExpression] = struct {
				StartTime metav1.Time
				EndTime   metav1.Time
			}{
				StartTime: schedule.StartTime,
				EndTime:   schedule.EndTime,
			}
		}

		for _, schedule := range schedules {
			hubSchedules = append(
				hubSchedules,
				hub.Schedule{
					CronExpression: schedule,
					StartTime:      remainderMap[schedule].StartTime,
					EndTime:        remainderMap[schedule].EndTime,
				},
			)
		}
		return hubSchedules
	}
	return hub.Triggers{
		Schedules:         convertSchedulesTo(triggers.Schedules, remainder),
		OnChange:          convertOnChangesTo(triggers.OnChange),
		RunConfigurations: triggers.RunConfigurations,
	}
}

func convertRuntimeParametersFrom(hubRtp []hub.RuntimeParameter) []RuntimeParameter {
	var rtp []RuntimeParameter
	for _, namedValue := range hubRtp {
		rtp = append(rtp, RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}
	return rtp
}

func convertArtifactLocatorFrom(al hub.ArtifactLocator) ArtifactLocator {
	return ArtifactLocator{
		Component: al.Component,
		Artifact:  al.Artifact,
		Index:     al.Index,
	}
}

func convertArtifactPathFrom(ap hub.ArtifactPath) ArtifactPath {
	return ArtifactPath{
		Locator: convertArtifactLocatorFrom(ap.Locator),
		Filter:  ap.Filter,
	}
}

func convertArtifactsFrom(hubArtifacts []hub.OutputArtifact) []OutputArtifact {
	var artifacts []OutputArtifact
	for _, artifact := range hubArtifacts {
		artifacts = append(artifacts, OutputArtifact{
			Name: artifact.Name,
			Path: convertArtifactPathFrom(artifact.Path),
		})
	}
	return artifacts
}

func convertTriggersFrom(triggers hub.Triggers) Triggers {
	convertSchedulesFrom := func(hubSchedules []hub.Schedule) (schedules []string) {
		for _, schedule := range hubSchedules {
			schedules = append(schedules, schedule.CronExpression)
		}
		return schedules
	}
	convertOnChangesFrom := func(hubOct []hub.OnChangeType) []OnChangeType {
		var oct []OnChangeType
		for _, onChange := range hubOct {
			oct = append(oct, OnChangeType(onChange))
		}
		return oct
	}
	return Triggers{
		Schedules:         convertSchedulesFrom(triggers.Schedules),
		OnChange:          convertOnChangesFrom(triggers.OnChange),
		RunConfigurations: triggers.RunConfigurations,
	}
}
