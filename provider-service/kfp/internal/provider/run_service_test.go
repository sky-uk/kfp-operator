//go:build unit

package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	latest "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v2"
)

var _ = Describe("RunService", func() {
	const (
		pipelineId        = "pipeline-id"
		experimentVersion = "experiment-version"
		pipelineVersionId = "pipeline-version-id"
	)

	var (
		mockClient mocks.MockRunServiceClient
		runService RunService
		rd         = testutil.RandomRunDefinition()
		ctx        = context.Background()
	)

	rd.Name.Name = "runName"
	rd.Name.Namespace = "runNamespace"
	rd.PipelineName.Name = "pipelineName"
	rd.PipelineName.Namespace = "pipelineNamespace"
	rd.RunConfigurationName.Name = "runConfigurationName"
	rd.RunConfigurationName.Namespace = "runConfigurationNamespace"
	rd.Artifacts = []latest.OutputArtifact{
		{Name: "artifact-name"},
		{Name: "artifact-name-2"},
	}

	rd.Parameters = map[string]string{
		"key-1": "value-1",
		"key-2": "value-2",
	}

	expectedRuntimeParams := map[string]*structpb.Value{
		"key-1": structpb.NewStringValue("value-1"),
		"key-2": structpb.NewStringValue("value-2"),
	}

	runAsDescription, err := yaml.Marshal(resource.References{
		RunName:              rd.Name,
		RunConfigurationName: rd.RunConfigurationName,
		PipelineName:         rd.PipelineName,
		Artifacts:            rd.Artifacts,
	})
	Expect(err).ToNot(HaveOccurred())

	expectedReq := &go_client.CreateRunRequest{
		Run: &go_client.Run{
			ExperimentId: experimentVersion,
			DisplayName:  "runNamespace-runName",
			Description:  string(runAsDescription),
			PipelineSource: &go_client.Run_PipelineVersionReference{
				PipelineVersionReference: &go_client.PipelineVersionReference{
					PipelineId:        pipelineId,
					PipelineVersionId: pipelineVersionId,
				},
			},
			RuntimeConfig: &go_client.RuntimeConfig{
				Parameters: expectedRuntimeParams,
			},
		},
	}

	BeforeEach(
		func() {
			mockClient = mocks.MockRunServiceClient{}
			runService = DefaultRunService{
				client:         &mockClient,
				labelGenerator: NoopLabelGen{},
			}
		},
	)

	Context("CreateRun", func() {
		It("should return a run id", func() {
			expectedId := "expected-id"
			mockClient.On("CreateRun", expectedReq).Return(
				&go_client.Run{RunId: expectedId},
				nil,
			)
			runId, err := runService.CreateRun(
				ctx,
				rd,
				pipelineId,
				pipelineVersionId,
				experimentVersion,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(runId).To(Equal(expectedId))
		})

		When("Name is invalid", func() {
			It("should return an error", func() {
				rdCopy := rd
				rdCopy.Name.Name = ""
				runId, err := runService.CreateRun(
					ctx,
					rdCopy,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(runId).To(BeEmpty())
				Expect(err).To(HaveOccurred())
			})
		})

		When("RunService Errors", func() {
			It("should return an error", func() {
				expectedErr := errors.New("error")
				mockClient.On("CreateRun", expectedReq).Return(nil, expectedErr)
				runId, err := runService.CreateRun(
					ctx,
					rd,
					pipelineId,
					pipelineVersionId,
					experimentVersion,
				)

				Expect(runId).To(BeEmpty())
				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
