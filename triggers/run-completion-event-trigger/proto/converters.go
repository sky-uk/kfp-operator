package run_completion_event_trigger

import "github.com/sky-uk/kfp-operator/argo/common"

func ProtoRunCompletionToCommon(protoRunCompletion *RunCompletionEvent) (common.RunCompletionEvent, error) {

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

	return common.RunCompletionEvent{
		Status:                StatusConverter(protoRunCompletion.Status),
		PipelineName:          pipelineName,
		RunConfigurationName:  &runConfigurationName,
		RunName:               &runName,
		RunId:                 protoRunCompletion.RunId,
		ServingModelArtifacts: ArtifactsConverter(protoRunCompletion.ServingModelArtifacts),
		Artifacts:             nil,
		Provider:              protoRunCompletion.Provider,
	}, nil
}

func ArtifactsConverter(artifacts []*ServingModelArtifact) []common.Artifact {
	var commonArtifacts []common.Artifact

	for _, pbArtifact := range artifacts {
		commonArtifact := common.Artifact{
			Name:     pbArtifact.Name,
			Location: pbArtifact.Location,
		}
		commonArtifacts = append(commonArtifacts, commonArtifact)
	}

	return commonArtifacts
}

func StatusConverter(status Status) common.RunCompletionStatus {
	switch status {
	case Status_SUCCEEDED:
		return common.RunCompletionStatuses.Succeeded
	default:
		return common.RunCompletionStatuses.Failed
	}
}
