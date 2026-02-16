package converters

import (
	"github.com/sky-uk/kfp-operator/pkg/common"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
	"time"
)

func ProtoRunCompletionToCommon(protoRunCompletion *pb.RunCompletionEvent) (common.RunCompletionEvent, error) {

	pipelineName, err := common.NamespacedNameFromString(protoRunCompletion.PipelineName)
	if err != nil {
		return common.RunCompletionEvent{}, err
	}

	runConfigurationName, err := common.NamespacedNameFromString(protoRunCompletion.RunConfigurationName)
	if err != nil {
		return common.RunCompletionEvent{}, err
	}

	runName, err := common.NamespacedNameFromString(protoRunCompletion.RunName)
	if err != nil {
		return common.RunCompletionEvent{}, err
	}

	provider, err := common.NamespacedNameFromString(protoRunCompletion.Provider)
	if err != nil {
		return common.RunCompletionEvent{}, err
	}

	var startTime *time.Time
	if protoRunCompletion.RunStartTime != nil {
		st := protoRunCompletion.RunStartTime.AsTime()
		startTime = &st
	}

	var endTime *time.Time
	if protoRunCompletion.RunEndTime != nil {
		et := protoRunCompletion.RunEndTime.AsTime()
		endTime = &et
	}

	return common.RunCompletionEvent{
		Status:                statusConverter(protoRunCompletion.Status),
		PipelineName:          pipelineName,
		RunConfigurationName:  &runConfigurationName,
		RunName:               &runName,
		RunId:                 protoRunCompletion.RunId,
		ServingModelArtifacts: protoToArtifacts(protoRunCompletion.ServingModelArtifacts),
		Artifacts:             protoToArtifacts(protoRunCompletion.Artifacts),
		Provider:              provider,
		RunStartTime:          startTime,
		RunEndTime:            endTime,
	}, nil
}

func protoToArtifacts(artifacts []*pb.Artifact) []common.Artifact {
	commonArtifacts := []common.Artifact{}

	for _, pbArtifact := range artifacts {
		commonArtifact := common.Artifact{
			Name:     pbArtifact.Name,
			Location: pbArtifact.Location,
		}
		commonArtifacts = append(commonArtifacts, commonArtifact)
	}

	return commonArtifacts
}

func statusConverter(status pb.Status) common.RunCompletionStatus {
	switch status {
	case pb.Status_SUCCEEDED:
		return common.RunCompletionStatuses.Succeeded
	default:
		return common.RunCompletionStatuses.Failed
	}
}
