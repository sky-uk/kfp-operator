//go:build unit
// +build unit

package kfp

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/kfp/ml_metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Context("gRPC Metadata Store", func() {
	var (
		mockCtrl                       *gomock.Controller
		mockMetadataStoreServiceClient *ml_metadata.MockMetadataStoreServiceClient
		store                          GrpcMetadataStore
		workflowName                   = RandomString()
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockMetadataStoreServiceClient = ml_metadata.NewMockMetadataStoreServiceClient(mockCtrl)
		store = GrpcMetadataStore{
			MetadataStoreServiceClient: mockMetadataStoreServiceClient,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("GetServingModelArtifact", func() {
		var givenArtifactTypeIdId = func() int64 {
			artifactId := RandomInt64()
			typeName := PushedModelArtifactType
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactType(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName})).
				Return(&ml_metadata.GetArtifactTypeResponse{
					ArtifactType: &ml_metadata.ArtifactType{Id: &artifactId},
				}, nil)

			return artifactId
		}

		var givenContextId = func() int64 {
			contextId := RandomInt64()
			typeName := PipelineRunTypeName
			mockMetadataStoreServiceClient.EXPECT().
				GetContextByTypeAndName(gomock.Any(), gomock.Eq(&ml_metadata.GetContextByTypeAndNameRequest{TypeName: &typeName, ContextName: &workflowName})).
				Return(&ml_metadata.GetContextByTypeAndNameResponse{
					Context: &ml_metadata.Context{
						Id: &contextId,
					},
				}, nil)

			return contextId
		}

		When("GetArtifactType errors with NotFound", func() {
			It("returns no artifacts", func() {
				typeName := PushedModelArtifactType
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactType(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName})).
					Return(nil, status.Error(codes.NotFound, "type not found"))

				servingModelArtifacts, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(servingModelArtifacts).To(BeEmpty())
			})
		})

		When("GetArtifactType errors", func() {
			It("Errors", func() {
				typeName := PushedModelArtifactType
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactType(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName})).
					Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactType returns an invalid artifact ID", func() {
			It("Errors", func() {
				typeName := PushedModelArtifactType
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactType(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName})).
					Return(&ml_metadata.GetArtifactTypeResponse{
						ArtifactType: &ml_metadata.ArtifactType{Id: nil},
					}, nil)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetContextByTypeAndName errors", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				typeName := PipelineRunTypeName
				mockMetadataStoreServiceClient.EXPECT().
					GetContextByTypeAndName(gomock.Any(), gomock.Eq(&ml_metadata.GetContextByTypeAndNameRequest{TypeName: &typeName, ContextName: &workflowName})).
					Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetContextByTypeAndName returns an invalid context ID", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				typeName := PipelineRunTypeName
				mockMetadataStoreServiceClient.EXPECT().
					GetContextByTypeAndName(gomock.Any(), gomock.Eq(&ml_metadata.GetContextByTypeAndNameRequest{TypeName: &typeName, ContextName: &workflowName})).
					Return(&ml_metadata.GetContextByTypeAndNameResponse{
						Context: nil,
					}, nil)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext errors", func() {
			It("Errors", func() {
				givenArtifactTypeIdId()
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(nil, fmt.Errorf("an error"))

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetArtifactsByContext does not return an artifact with the correct type", func() {
			It("filters out invalid artifacts", func() {
				givenArtifactTypeIdId()
				anotherArtifactTypeId := RandomInt64()
				artifactLocation := RandomString()
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{},
							{
								TypeId: &anotherArtifactTypeId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: RandomString(),
										},
									},
								},
							},
						},
					}, nil)

				_, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("GetArtifactsByContext returns an artifact with a non-string 'name' property", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
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
					}, nil)

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
				artifactLocation := RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
							},
						},
					}, nil)

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
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
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
					}, nil)

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
				artifactLocation := RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
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
											IntValue: RandomExceptOne(),
										},
									},
								},
							},
						},
					}, nil)

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
				firstArtifactLocation := RandomString()
				secondArtifactLocation := RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
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
					}, nil)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(ContainElements(
					ServingModelArtifact{
						Name:     "first-model",
						Location: firstArtifactLocation,
					},
					ServingModelArtifact{
						Name:     "second-model",
						Location: secondArtifactLocation,
					},
				))
			})
		})
	})
})
