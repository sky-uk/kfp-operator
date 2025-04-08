package webhook

import (
	"github.com/hashicorp/go-bexpr"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type EventProcessor interface {
	ToRunCompletionEvent(eventData *common.RunCompletionEventData, runConfiguration *pipelineshub.RunConfiguration, run *pipelineshub.Run) (*common.RunCompletionEvent, EventError)
}

type FilterFunc func([]common.PipelineComponent, []pipelineshub.OutputArtifact) []common.Artifact

type ResourceArtifactsEventProcessor struct {
	filter FilterFunc
}

func NewResourceArtifactsEventProcessor() EventProcessor {
	return ResourceArtifactsEventProcessor{filter: filterByResourceArtifacts}
}

func (ep ResourceArtifactsEventProcessor) ToRunCompletionEvent(eventData *common.RunCompletionEventData, runConfiguration *pipelineshub.RunConfiguration, run *pipelineshub.Run) (*common.RunCompletionEvent, EventError) {
	artifacts := []pipelineshub.OutputArtifact{}
	if runConfiguration != nil {
		artifacts = runConfiguration.Spec.Run.Artifacts
	} else if run != nil {
		artifacts = run.Spec.Artifacts
	} else {
		return nil, &InvalidEvent{Msg: "no RunConfiguration or RunName specified"}
	}

	runCompletionEvent := eventData.ToRunCompletionEvent()
	runCompletionEvent.Artifacts = ep.filter(eventData.PipelineComponents, artifacts)

	return &runCompletionEvent, nil
}

func filterByResourceArtifacts(pipelineComponents []common.PipelineComponent, outputArtifacts []pipelineshub.OutputArtifact) []common.Artifact {
	artifacts := make([]common.Artifact, 0)
	for _, outputArtifact := range outputArtifacts {
		var evaluator *bexpr.Evaluator
		var err error

		if outputArtifact.Path.Filter != "" {
			evaluator, err = bexpr.CreateEvaluator(outputArtifact.Path.Filter)
			if err != nil {
				continue
			}
		}

		for _, component := range pipelineComponents {
			if component.Name != outputArtifact.Path.Locator.Component {
				continue
			}

			for _, componentArtifact := range component.ComponentArtifacts {
				if componentArtifact.Name != outputArtifact.Path.Locator.Artifact {
					continue
				}
				if outputArtifact.Path.Locator.Index >= len(componentArtifact.Artifacts) {
					continue
				}
				artifact := componentArtifact.Artifacts[outputArtifact.Path.Locator.Index]
				if artifact.Uri == "" {
					continue
				}

				if evaluator != nil {
					matched, err := evaluator.Evaluate(artifact.Metadata)
					// evaluator errors on missing properties
					if err != nil || !matched {
						continue
					}
				}

				artifacts = append(artifacts, common.Artifact{Name: outputArtifact.Name, Location: artifact.Uri})
			}
		}
	}
	return artifacts
}
