//go:build unit
// +build unit

package kfp

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/kfp/ml_metadata"
)

var _ = Context("gRPC Metadata Store", func() {
	var (
		mockCtrl                       *gomock.Controller
		mockMetadataStoreServiceClient *ml_metadata.MockMetadataStoreServiceClient
		store                          GrpcMetadataStore
		workflowName                   = common.RandomString()
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
			contextId := common.RandomInt64()
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

		When("GetContextByTypeAndName errors", func() {
			It("Errors", func() {
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
				anotherArtifactTypeId := common.RandomInt64()
				artifactLocation := common.RandomString()
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
											StringValue: common.RandomString(),
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
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Uri: &artifactLocation,
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
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Uri: &artifactLocation,
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
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "Pusher:pushed_model:0",
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
				contextId := givenContextId()
				artifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Uri: &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "Pusher:pushed_model:0",
										},
									},
								},
							},
							{
								Uri: &artifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "Pusher:pushed_model:0",
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
					}, nil)

				results, err := store.GetServingModelArtifact(context.Background(), workflowName)
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
				Expect(results).To(Not(BeNil()))
			})
		})

		When("GetArtifactsByContext returns valid artifacts", func() {
			It("Returns all ServingModelLocations", func() {
				contextId := givenContextId()
				firstArtifactLocation := common.RandomString()
				secondArtifactLocation := common.RandomString()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								Uri: &firstArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "Pusher:pushed_model:0",
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
								Uri: &secondArtifactLocation,
								CustomProperties: map[string]*ml_metadata.Value{
									ArtifactNameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "Pusher:pushed_model:0",
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
					common.Artifact{
						Name:     "Pusher:pushed_model:0",
						Location: firstArtifactLocation,
					},
					common.Artifact{
						Name:     "Pusher:pushed_model:0",
						Location: secondArtifactLocation,
					},
				))
			})
		})
	})
})
