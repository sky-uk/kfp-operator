package v1alpha5

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)
	remainder := RunScheduleConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsTo(src.Spec.Artifacts)
	dst.Spec.Schedule = convertScheduleTo(
		src.Spec.Schedule,
		remainder.Schedule,
	)
	if err := pipelines.TransformInto(src.Status, &dst.Status); err != nil {
		return err
	}
	namespacedName := convertProviderTo(remainder.Provider)
	dst.Spec.Provider = namespacedName
	dst.Status.Provider = convertProviderAndIdTo(
		src.Status.ProviderId,
		namespacedName.Namespace,
	)

	return nil
}

func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)
	remainder := RunScheduleConversionRemainder{}
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Artifacts = convertArtifactsFrom(src.Spec.Artifacts)
	schedule, err := convertCronExpressionFrom(src.Spec.Schedule, &remainder)
	if err != nil {
		return err
	}
	dst.Spec.Schedule = schedule
	if err := pipelines.TransformInto(src.Status, &dst.Status); err != nil {
		return err
	}
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)
	remainder.Provider = src.Status.Provider.Name

	return pipelines.SetConversionAnnotations(dst, &remainder)
}

// +kubebuilder:object:generate=false
type ConversionError struct {
	Message string
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("Error during conversion: %s", e.Message)
}

func convertCronExpressionFrom(
	schedule hub.Schedule,
	remainder *RunScheduleConversionRemainder,
) (string, error) {
	if remainder == nil {
		return "", &ConversionError{"expected a v1alpha6 remainder but got nil"}
	}
	remainder.Schedule = schedule
	return schedule.CronExpression, nil
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
