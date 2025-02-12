//go:build unit || decoupled

package webhook

import (
	"context"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/common"
)

type StubbedEventProcessor struct {
	expectedRunCompletionEventData *common.RunCompletionEventData
	returnedRunCompletionEvent     *common.RunCompletionEvent
	expectedError                  error
}

func (sep StubbedEventProcessor) ToRunCompletionEvent(_ context.Context, passedData common.RunCompletionEventData) (*common.RunCompletionEvent, error) {
	if sep.expectedRunCompletionEventData != nil {
		Expect(passedData).To(Equal(*sep.expectedRunCompletionEventData))
	}
	return sep.returnedRunCompletionEvent, sep.expectedError
}

func randomComponentArtifactInstance() common.ComponentArtifactInstance {
	return common.ComponentArtifactInstance{
		Uri: common.RandomString(),
		Metadata: map[string]interface{}{
			"x": map[string]interface{}{
				"y": float64(1),
			},
			"pushed":             float64(1),
			"pushed_destination": "gs://somebucket",
		},
	}
}

func randomComponentArtifact() common.ComponentArtifact {
	return common.ComponentArtifact{
		Name:      common.RandomString(),
		Artifacts: apis.RandomNonEmptyList(randomComponentArtifactInstance),
	}
}

func randomPipelineComponent() common.PipelineComponent {
	return common.PipelineComponent{
		Name:               common.RandomString(),
		ComponentArtifacts: apis.RandomNonEmptyList(randomComponentArtifact),
	}
}

func RandomRunCompletionEventData() common.RunCompletionEventData {
	runName := common.RandomNamespacedName()
	runConfigurationName := common.RandomNamespacedName()

	return common.RunCompletionEventData{
		Status:                common.RunCompletionStatuses.Succeeded,
		PipelineName:          common.NamespacedName{},
		RunConfigurationName:  &runConfigurationName,
		RunName:               &runName,
		RunId:                 common.RandomString(),
		ServingModelArtifacts: apis.RandomNonEmptyList(common.RandomArtifact),
		PipelineComponents:    apis.RandomNonEmptyList(randomPipelineComponent),
		Provider:              common.RandomString(),
	}
}
