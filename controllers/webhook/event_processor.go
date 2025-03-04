package webhook

import (
	"context"
	"errors"

	"github.com/hashicorp/go-bexpr"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1beta1"
	"github.com/sky-uk/kfp-operator/argo/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EventProcessor interface {
	ToRunCompletionEvent(ctx context.Context, runData common.RunCompletionEventData) (*common.RunCompletionEvent, error)
}

type FilterFunc func([]common.PipelineComponent, []pipelinesv1.OutputArtifact) []common.Artifact

type ResourceArtifactsEventProcessor struct {
	client client.Reader
	filter FilterFunc
}

func NewResourceArtifactsEventProcessor(client client.Reader) EventProcessor {
	return ResourceArtifactsEventProcessor{client: client, filter: filterByResourceArtifacts}
}

func (ep ResourceArtifactsEventProcessor) ToRunCompletionEvent(ctx context.Context, runData common.RunCompletionEventData) (*common.RunCompletionEvent, error) {
	outputArtifacts, err := extractResourceArtifacts(ctx, ep.client, runData.RunConfigurationName, runData.RunName)
	if err != nil {
		return nil, err
	}

	runCompletionEvent := runData.ToRunCompletionEvent()
	runCompletionEvent.Artifacts = ep.filter(runData.PipelineComponents, outputArtifacts)

	return &runCompletionEvent, nil
}

func filterByResourceArtifacts(pipelineComponents []common.PipelineComponent, outputArtifacts []pipelinesv1.OutputArtifact) []common.Artifact {
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

func extractResourceArtifacts(ctx context.Context, reader client.Reader, runConfigurationName *common.NamespacedName, runName *common.NamespacedName) ([]pipelinesv1.OutputArtifact, error) {
	logger := common.LoggerFromContext(ctx)
	if runConfigurationName != nil {
		runConfigurationResource := &pipelinesv1.RunConfiguration{}
		if err := reader.Get(ctx, client.ObjectKey{
			Namespace: runConfigurationName.Namespace,
			Name:      runConfigurationName.Name,
		}, runConfigurationResource); err != nil {
			logger.Error(err, "failed to load RunConfiguration", "RunConfig", runConfigurationName)
			return nil, err
		}
		return runConfigurationResource.Spec.Run.Artifacts, nil
	} else if runName != nil {
		runResource := &pipelinesv1.Run{}
		if err := reader.Get(ctx, client.ObjectKey{
			Namespace: runName.Namespace,
			Name:      runName.Name,
		}, runResource); err != nil {
			logger.Error(err, "failed to load Run", "Run", runName)
			return nil, err
		}
		return runResource.Spec.Artifacts, nil
	} else {
		err := errors.New("no RunConfiguration or RunName specified")
		logger.Error(err, "failed to retrieve resource artifacts")
		return nil, err
	}
}
