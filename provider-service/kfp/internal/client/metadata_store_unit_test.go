//go:build unit

package client

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Context("gRPC Metadata Store", func() {
	var (
		mockMetadataStoreServiceClient mocks.MockMetadataStoreServiceClient
		store                          GrpcMetadataStore
		workflowName                   = common.RandomString()
	)

	BeforeEach(func() {
		mockMetadataStoreServiceClient = mocks.MockMetadataStoreServiceClient{}
		store = GrpcMetadataStore{
			MetadataStoreServiceClient: &mockMetadataStoreServiceClient,
		}
	})

	var givenContextId = func() int64 {
		contextId := common.RandomInt64()
		typeName := PipelineRunTypeName
		mockMetadataStoreServiceClient.On(
			"GetContextByTypeAndName",
			&ml_metadata.GetContextByTypeAndNameRequest{
				TypeName:    &typeName,
				ContextName: &workflowName,
			},
		).Return(
			&ml_metadata.GetContextByTypeAndNameResponse{
				Context: &ml_metadata.Context{
					Id: &contextId,
				},
			},
			nil,
		)
		return contextId
	}

	Describe("GetArtifactsForRun", func() {
		var (
			artifactName = common.RandomString()
			contextId    = int64(123)
			executionId  = int64(456)
			artifactId   = int64(789)
		)

		It("returns error if GetContextByTypeAndName fails", func() {
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(nil, fmt.Errorf("context error"))

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if context is nil", func() {
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: nil}, nil)

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if GetExecutionsByContext fails", func() {
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: &ml_metadata.Context{Id: &contextId}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetExecutionsByContext",
				mock.Anything,
				mock.Anything,
			).Return(nil, fmt.Errorf("executions error"))

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if GetEventsByExecutionIDs fails", func() {
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: &ml_metadata.Context{Id: &contextId}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetExecutionsByContext",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetExecutionsByContextResponse{Executions: []*ml_metadata.Execution{{Id: &executionId}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetEventsByExecutionIDs",
				mock.Anything,
				mock.Anything,
			).Return(nil, fmt.Errorf("events error"))

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if GetArtifactsByID fails", func() {
			eventType := ml_metadata.Event_OUTPUT
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: &ml_metadata.Context{Id: &contextId}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetExecutionsByContext",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetExecutionsByContextResponse{Executions: []*ml_metadata.Execution{{Id: &executionId}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetEventsByExecutionIDs",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetEventsByExecutionIDsResponse{Events: []*ml_metadata.Event{{
				Type:        &eventType,
				ExecutionId: &executionId,
				ArtifactId:  &artifactId,
			}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetArtifactsByID",
				mock.Anything,
				mock.Anything,
			).Return(nil, fmt.Errorf("artifacts error"))

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).To(HaveOccurred())
		})

		It("returns empty if no artifacts found", func() {
			eventType := ml_metadata.Event_OUTPUT
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: &ml_metadata.Context{Id: &contextId}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetExecutionsByContext",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetExecutionsByContextResponse{Executions: []*ml_metadata.Execution{{Id: &executionId}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetEventsByExecutionIDs",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetEventsByExecutionIDsResponse{Events: []*ml_metadata.Event{{
				Type:        &eventType,
				ExecutionId: &executionId,
				ArtifactId:  &artifactId,
			}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetArtifactsByID",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetArtifactsByIDResponse{Artifacts: []*ml_metadata.Artifact{}}, nil)

			results, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("returns artifacts for successful flow", func() {
			eventType := ml_metadata.Event_OUTPUT
			artifactUri := common.RandomString()
			mockMetadataStoreServiceClient.On(
				"GetContextByTypeAndName",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetContextByTypeAndNameResponse{Context: &ml_metadata.Context{Id: &contextId}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetExecutionsByContext",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetExecutionsByContextResponse{Executions: []*ml_metadata.Execution{{Id: &executionId}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetEventsByExecutionIDs",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetEventsByExecutionIDsResponse{Events: []*ml_metadata.Event{{
				Type:        &eventType,
				ExecutionId: &executionId,
				ArtifactId:  &artifactId,
			}}}, nil)
			mockMetadataStoreServiceClient.On(
				"GetArtifactsByID",
				mock.Anything,
				mock.Anything,
			).Return(&ml_metadata.GetArtifactsByIDResponse{Artifacts: []*ml_metadata.Artifact{{
				Id:   &artifactId,
				Name: &artifactName,
				Uri:  &artifactUri,
			}}}, nil)

			_, err := store.GetArtifactsForRun(context.Background(), workflowName)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Legacy: GetServingModelArtifacts", func() {
		var givenArtifactTypeIdId = func() int64 {
			artifactId := common.RandomInt64()
			typeName := PushedModelArtifactType

			mockMetadataStoreServiceClient.On(
				"GetArtifactType",
				&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName},
			).Return(
				&ml_metadata.GetArtifactTypeResponse{
					ArtifactType: &ml_metadata.ArtifactType{Id: &artifactId},
				},
				nil,
			)

			return artifactId
		}

		When("GetArtifactType errors with NotFound", func() {
			It("returns no artifacts", func() {
				typeName := PushedModelArtifactType

				mockMetadataStoreServiceClient.On(
					"GetArtifactType",
					&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName},
				).Return(nil, status.Error(codes.NotFound, "type not found"))

				servingModelArtifacts, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(servingModelArtifacts).To(BeEmpty())
			})
		})

		When("GetArtifactType errors", func() {
			It("Errors", func() {
				typeName := PushedModelArtifactType

				mockMetadataStoreServiceClient.On(
					"GetArtifactType",
					&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName},
				).Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactType returns an invalid artifact ID", func() {
			It("Errors", func() {
				typeName := PushedModelArtifactType
				mockMetadataStoreServiceClient.On(
					"GetArtifactType",
					&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName},
				).Return(
					&ml_metadata.GetArtifactTypeResponse{
						ArtifactType: &ml_metadata.ArtifactType{Id: nil},
					},
					nil,
				)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetContextByTypeAndName errors", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				typeName := PipelineRunTypeName

				mockMetadataStoreServiceClient.On(
					"GetContextByTypeAndName",
					&ml_metadata.GetContextByTypeAndNameRequest{
						TypeName:    &typeName,
						ContextName: &workflowName,
					},
				).Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetContextByTypeAndName returns an invalid context ID", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				typeName := PipelineRunTypeName
				mockMetadataStoreServiceClient.On(
					"GetContextByTypeAndName",
					&ml_metadata.GetContextByTypeAndNameRequest{
						TypeName:    &typeName,
						ContextName: &workflowName,
					},
				).Return(
					&ml_metadata.GetContextByTypeAndNameResponse{
						Context: nil,
					},
					nil,
				)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext errors", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				contextId := givenContextId()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext does not return an artifact with the correct type", func() {
			It("filters out invalid artifacts", func() {
				givenArtifactTypeIdId()
				anotherArtifactTypeId := common.RandomInt64()
				artifactLocation := common.RandomString()
				contextId := givenContextId()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{},
							{
								TypeId: &anotherArtifactTypeId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: common.RandomString(),
										},
									},
								},
							},
						},
					},
					nil,
				)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("GetArtifactsByContext returns an artifact with a non-string 'name' property", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_IntValue{
											IntValue: 42,
										},
									},
								},
							},
						},
					},
					nil,
				)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(results).To(Not(BeNil()))
			})
		})

		When("GetArtifactsByContext returns an artifact with no name", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
							},
						},
					},
					nil,
				)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(results).To(Not(BeNil()))
			})
		})

		When("GetArtifactsByContext returns an artifact with no uri", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
								},
							},
						},
					},
					nil,
				)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(results).To(Not(BeNil()))
			})
		})

		When("GetArtifactsByContext returns an artifact without the 'pushed' flag being true", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
								},
							},
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
									PushedCustomProperty: {
										Value: &ml_metadata.Value_IntValue{
											IntValue: common.RandomExceptOne(),
										},
									},
								},
							},
						},
					},
					nil,
				)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(results).To(Not(BeNil()))
			})
		})

		When("GetArtifactsByContext returns valid artifacts", func() {
			It("Returns all ServingModelLocations", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				firstArtifactLocation := common.RandomString()
				secondArtifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &firstArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
									PushedCustomProperty: {
										Value: &ml_metadata.Value_IntValue{
											IntValue: 1,
										},
									},
								},
							},
							{
								TypeId: &artifactId,
								Uri:    &secondArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "second-model",
										},
									},
									PushedCustomProperty: {
										Value: &ml_metadata.Value_IntValue{
											IntValue: 1,
										},
									},
								},
							},
						},
					},
					nil,
				)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(ContainElements(
					common.Artifact{
						Name:     "first-model",
						Location: firstArtifactLocation,
					},
					common.Artifact{
						Name:     "second-model",
						Location: secondArtifactLocation,
					},
				))
			})
		})
	})
})
