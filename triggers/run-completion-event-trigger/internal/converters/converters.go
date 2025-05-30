package converters

import (
	"github.com/sky-uk/kfp-operator/argo/common"
	pb "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/proto"
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

	var training *common.Training
	if protoRunCompletion.Training != nil {
		training = &common.Training{}
		if protoRunCompletion.Training.StartTime != nil {
			st := protoRunCompletion.Training.StartTime.AsTime()
			training.StartTime = &st
		}
		if protoRunCompletion.Training.EndTime != nil {
			et := protoRunCompletion.Training.EndTime.AsTime()
			training.EndTime = &et
		}
	}

	return common.RunCompletionEvent{
		Status:                statusConverter(protoRunCompletion.Status),
		PipelineName:          pipelineName,
		RunConfigurationName:  &runConfigurationName,
		RunName:               &runName,
		RunId:                 protoRunCompletion.RunId,
		ServingModelArtifacts: protoToArtifacts(protoRunCompletion.ServingModelArtifacts),
		Artifacts:             protoToArtifacts(protoRunCompletion.Artifacts),
		Provider:              protoRunCompletion.Provider,
		Training:              training,
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
