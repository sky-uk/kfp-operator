//go:build unit

package provider

import (
	"context"
	"errors"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	latest "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"gopkg.in/yaml.v2"
)

var _ = Describe("RunService", func() {
	const (
		pipelineId        = "pipeline-id"
		experimentVersion = "experiment-version"
		pipelineVersionId = "pipeline-version-id"
	)

	var (
		mockRunServiceClient mocks.MockRunServiceClient
		runService           RunService
		rd                   = testutil.RandomRunDefinition()
		ctx                  = context.Background()
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

	expectedRuntimeParams := []*go_client.Parameter{
		{Name: "key-1", Value: "value-1"},
		{Name: "key-2", Value: "value-2"},
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
			Name:        "runNamespace-runName",
			Description: string(runAsDescription),
			PipelineSpec: &go_client.PipelineSpec{
				PipelineId: pipelineId,
				Parameters: expectedRuntimeParams,
			},
			ResourceReferences: []*go_client.ResourceReference{
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_EXPERIMENT,
						Id:   experimentVersion,
					},
					Relationship: go_client.Relationship_OWNER,
				},
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_PIPELINE_VERSION,
						Id:   pipelineVersionId,
					},
					Relationship: go_client.Relationship_CREATOR,
				},
				{
					Key: &go_client.ResourceKey{
						Type: go_client.ResourceType_NAMESPACE,
						Id:   rd.Name.Namespace,
					},
					Relationship: go_client.Relationship_OWNER,
				},
			},
		},
	}

	BeforeEach(
		func() {
			mockRunServiceClient = mocks.MockRunServiceClient{}
			runService = DefaultRunService{
				client: &mockRunServiceClient,
			}
		},
	)

	Context("CreateRun", func() {
		It("should return a run id", func() {
			expectedId := "expected-id"
			mockRunServiceClient.On("CreateRun", expectedReq).Return(
				&go_client.RunDetail{Run: &go_client.Run{Id: expectedId}},
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
				mockRunServiceClient.On("CreateRun", expectedReq).Return(nil, expectedErr)
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
