package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func convertScheduleToHubWithRemainder(schedule string, remainder hub.Schedule) (hubSchedule hub.Schedule) {
	return hub.Schedule{
		CronExpression: schedule,
		StartTime:      remainder.StartTime,
		EndTime:        remainder.EndTime,
	}
}

// v1alpha5 -> v1alpha6
func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)
	// get old schedules out from remainder
	v1alpha6remainder := hub.RunScheduleConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha6remainder); err != nil {
		return err
	}
	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsToHub(src.Spec.Artifacts)

	dst.Spec.Schedule = convertScheduleToHubWithRemainder(src.Spec.Schedule, v1alpha6remainder.Schedule)
	dst.Status = hub.Status{
		ProviderId: hub.ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           hub.Conditions(src.Status.Conditions),
	}

	return nil
}

// v1alpha6 -> v1alpha5
func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)
	v1alpha6remainder := hub.RunScheduleConversionRemainder{}

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsFromHub(src.Spec.Artifacts)
	dst.Spec.Schedule = src.Spec.Schedule.CronExpression
	v1alpha6remainder.Schedule = src.Spec.Schedule
	dst.Status = Status{
		ProviderId: ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           Conditions(src.Status.Conditions),
	}

	return pipelines.SetConversionAnnotations(dst, &v1alpha6remainder)
}
