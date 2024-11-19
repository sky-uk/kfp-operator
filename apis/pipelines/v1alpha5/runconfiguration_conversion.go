package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	v1alpha6Remainder := hub.RunConfigurationConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &v1alpha6Remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Provider = getProviderAnnotation(src)
	removeProviderAnnotation(dst)
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersTo(
		src.Spec.Run.RuntimeParameters,
	)
	dst.Spec.Run.Artifacts = convertArtifactsTo(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersTo(src.Spec.Triggers, v1alpha6Remainder)

	if err := pipelines.TransformInto(src.Status, &dst.Status); err != nil {
		return err
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)
	v1alpha6Remainder := hub.RunConfigurationConversionRemainder{}
	dst.ObjectMeta = src.ObjectMeta
	setProviderAnnotation(src.Spec.Run.Provider, &dst.ObjectMeta)
	dst.Spec.Run.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersFrom(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsFrom(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersFrom(src.Spec.Triggers, &v1alpha6Remainder)

	if err := pipelines.TransformInto(src.Status, &dst.Status); err != nil {
		return err
	}

	return pipelines.SetConversionAnnotations(dst, &v1alpha6Remainder)
}

// Converts spoke Triggers into hub Triggers whilst taking into account of
// annotations that convey hub Triggers (remainder).
//
// If the spoke Schedule matches a CronExpression in the remainder then the
// conversion will use the StartTime and EndTime pointers from the remainder.
// If the spoke Schedules does not have a matching CronExpression in the
// remainder then the StartTime and EndTime pointers will be set to nil.
func convertTriggersTo(
	triggers Triggers,
	remainder hub.RunConfigurationConversionRemainder,
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
		remainder hub.RunConfigurationConversionRemainder,
	) []hub.Schedule {
		// map of the hub CronExpression -> { StartTime, EndTime }.
		// This could potentially be lossy because if two schedules share
		// the same CronExpression, then one of them will be overwritten.
		remainderMap := make(
			map[string]struct {
				StartTime *metav1.Time
				EndTime   *metav1.Time
			},
		)
		for _, schedule := range remainder.Schedules {
			remainderMap[schedule.CronExpression] = struct {
				StartTime *metav1.Time
				EndTime   *metav1.Time
			}{
				StartTime: schedule.StartTime,
				EndTime:   schedule.EndTime,
			}
		}
		var hubSchedules []hub.Schedule
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

func convertTriggersFrom(
	triggers hub.Triggers,
	remainder *hub.RunConfigurationConversionRemainder,
) Triggers {
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
	var remainderSchedules []hub.Schedule
	for _, schedule := range triggers.Schedules {
		if schedule.StartTime != nil || schedule.EndTime != nil {
			remainderSchedules = append(remainderSchedules, schedule)
		}
	}
	if len(remainderSchedules) > 0 {
		remainder.Schedules = remainderSchedules
	}
	return Triggers{
		Schedules:         convertSchedulesFrom(triggers.Schedules),
		OnChange:          convertOnChangesFrom(triggers.OnChange),
		RunConfigurations: triggers.RunConfigurations,
	}
}

func (v *ValueFrom) convertToHub() *hub.ValueFrom {
	if v != nil {
		return &hub.ValueFrom{
			RunConfigurationRef: hub.RunConfigurationRef{
				Name:           v.RunConfigurationRef.Name,
				OutputArtifact: v.RunConfigurationRef.OutputArtifact,
			},
		}
	}
	return nil
}

func convertFromHubValueFrom(v *hub.ValueFrom) *ValueFrom {
	if v != nil {
		return &ValueFrom{
			RunConfigurationRef: RunConfigurationRef{
				Name:           v.RunConfigurationRef.Name,
				OutputArtifact: v.RunConfigurationRef.OutputArtifact,
			},
		}
	}
	return nil
}

func convertRuntimeParametersTo(rtp []RuntimeParameter) []hub.RuntimeParameter {
	var hubRtp []hub.RuntimeParameter
	for _, namedValue := range rtp {
		hubRtp = append(hubRtp, hub.RuntimeParameter{
			Name:      namedValue.Name,
			Value:     namedValue.Value,
			ValueFrom: namedValue.ValueFrom.convertToHub(),
		})
	}
	return hubRtp
}

func convertRuntimeParametersFrom(hubRtp []hub.RuntimeParameter) []RuntimeParameter {
	var rtp []RuntimeParameter
	for _, namedValue := range hubRtp {
		rtp = append(rtp, RuntimeParameter{
			Name:      namedValue.Name,
			Value:     namedValue.Value,
			ValueFrom: convertFromHubValueFrom(namedValue.ValueFrom),
		})
	}
	return rtp
}
