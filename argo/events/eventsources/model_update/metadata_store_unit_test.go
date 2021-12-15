//go:build unit
// +build unit

package model_update

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/rand"
	"pipelines.kubeflow.org/events/ml_metadata"
)

var _ = Context("gRPC Metadata Store", func() {
	var (
		mockCtrl                       *gomock.Controller
		mockMetadataStoreServiceClient *ml_metadata.MockMetadataStoreServiceClient
		store                          GrpcMetadataStore
		workflowName                   = randomString()
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
		var givenContextId = func() int64 {
			contextId := int64(rand.Int())
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

		var givenArtifactTypeIdId = func() int64 {
			artifactId := int64(rand.Int())
			typeName := PushedModelArtifactType
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactType(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactTypeRequest{TypeName: &typeName})).
				Return(&ml_metadata.GetArtifactTypeResponse{
					ArtifactType: &ml_metadata.ArtifactType{Id: &artifactId},
				}, nil)

			return artifactId
		}

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

		When("GetArtifactsByContext does not return an artifact with the property", func() {
			It("filters out invalid artifacts", func() {
				givenArtifactTypeIdId()
				anotherArtifactTypeId := int64(rand.Int())
				artifactLocation := randomString()
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
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: randomString(),
										},
									},
								},
							},
						},
					}, nil)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		When("GetArtifactsByContext returns Artifacts with non-string 'name' properties", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := randomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
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
			})
		})

		When("GetArtifactsByContext returns artifacts with the property but no name", func() {
			It("filters out invalid artifacts", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				artifactLocation := randomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{},
							{
								TypeId: &artifactId,
								Uri:    &artifactLocation,
							},
						},
					}, nil)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})

		When("GetArtifactsByContext returns artifacts with the property", func() {
			It("Returns all property values", func() {
				artifactId := givenArtifactTypeIdId()
				contextId := givenContextId()
				firstArtifactLocation := randomString()
				secondArtifactLocation := randomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								TypeId: &artifactId,
								Uri:    &firstArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
								},
							},
							{
								TypeId: &artifactId,
								Uri:    &secondArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "second-model",
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
