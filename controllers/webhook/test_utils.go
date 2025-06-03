//go:build unit || decoupled

package webhook

import (
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"time"
)

var StaticTime = time.Now().Round(0)

type StubbedEventProcessor struct {
	expectedRunCompletionEventData *common.RunCompletionEventData
	returnedRunCompletionEvent     *common.RunCompletionEvent
	expectedError                  EventError
}

func (sep StubbedEventProcessor) ToRunCompletionEvent(eventData *common.RunCompletionEventData, runConfiguration *pipelineshub.RunConfiguration, run *pipelineshub.Run) (*common.RunCompletionEvent, EventError) {
	if sep.expectedRunCompletionEventData != nil {
		Expect(eventData).To(Equal(sep.expectedRunCompletionEventData))
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
		RunStartTime:          &StaticTime,
		RunEndTime:            &StaticTime,
		ServingModelArtifacts: apis.RandomNonEmptyList(common.RandomArtifact),
		PipelineComponents:    apis.RandomNonEmptyList(randomPipelineComponent),
		Provider:              common.RandomString(),
	}
}
