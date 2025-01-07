//go:build unit

package client

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/ml_metadata"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/utils/pointer"
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

	Describe("GetArtifacts", func() {
		artifactName := common.RandomString()
		componentName := common.RandomString()
		outputName := common.RandomString()
		artifactDefs := []pipelinesv1.OutputArtifact{{
			Name: artifactName,
			Path: pipelinesv1.ArtifactPath{
				Locator: pipelinesv1.ArtifactLocator{
					Component: componentName,
					Artifact:  outputName,
					Index:     1,
				},
				Filter: "x.y == 1",
			},
		}}
		artifactPath := fmt.Sprintf("%s:%s:1", componentName, outputName)

		When("GetContextByTypeAndName errors", func() {
			It("Errors", func() {
				typeName := PipelineRunTypeName

				mockMetadataStoreServiceClient.On(
					"GetContextByTypeAndName",
					&ml_metadata.GetContextByTypeAndNameRequest{
						TypeName:    &typeName,
						ContextName: &workflowName,
					},
				).Return(nil, fmt.Errorf("an error"))

				_, err := store.GetArtifacts(context.Background(), workflowName, []pipelinesv1.OutputArtifact{})
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetContextByTypeAndName returns an invalid context ID", func() {
			It("Errors", func() {
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

				_, err := store.GetArtifacts(context.Background(), workflowName, []pipelinesv1.OutputArtifact{})
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext errors", func() {
			It("Errors", func() {
				contextId := givenContextId()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(nil, fmt.Errorf("an error"))

				_, err := store.GetArtifacts(context.Background(), workflowName, []pipelinesv1.OutputArtifact{})
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext returns artifacts", func() {
			validProperties := map[string]*ml_metadata.Value{
				"x": {
					Value: &ml_metadata.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"y": structpb.NewNumberValue(1),
							},
						},
					},
				},
			}

			It("filters out artifacts that don't match the name", func() {
				contextId := givenContextId()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Name:             pointer.String(common.RandomString()),
								Uri:              pointer.String(apis.RandomString()),
								CustomProperties: validProperties,
							},
						},
					},
					nil,
				)

				results, err := store.GetArtifacts(context.Background(), workflowName, artifactDefs)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})

			It("filters out artifacts that doesn't have a URI", func() {
				contextId := givenContextId()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Name:             pointer.String(artifactPath),
								CustomProperties: validProperties,
							},
						},
					},
					nil,
				)

				results, err := store.GetArtifacts(context.Background(), workflowName, artifactDefs)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})

			It("filters out artifacts that misses properties", func() {
				contextId := givenContextId()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Name: pointer.String(artifactPath),
								Uri:  pointer.String(artifactPath),
							},
						},
					},
					nil,
				)

				results, err := store.GetArtifacts(context.Background(), workflowName, artifactDefs)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})

			It("filters out artifacts that has properties that don't match", func() {
				contextId := givenContextId()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Name: pointer.String(artifactPath),
								Uri:  pointer.String(common.RandomString()),
								CustomProperties: map[string]*ml_metadata.Value{
									common.RandomString(): {
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

				results, err := store.GetArtifacts(context.Background(), workflowName, artifactDefs)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})

			It("returns matching artifacts", func() {
				contextId := givenContextId()
				artifactLocation := common.RandomString()

				mockMetadataStoreServiceClient.On(
					"GetArtifactsByContext",
					&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId},
				).Return(
					&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Name:             pointer.String(artifactPath),
								Uri:              &artifactLocation,
								CustomProperties: validProperties,
							},
						},
					},
					nil,
				)

				results, err := store.GetArtifacts(context.Background(), workflowName, artifactDefs)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(ContainElements(
					common.Artifact{
						Name:     artifactName,
						Location: artifactLocation,
					},
				))
			})
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
