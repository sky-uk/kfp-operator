//go:build unit
// +build unit

package main

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/util/rand"
	"pipelines.kubeflow.org/events/ml_metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("gRPC Metadata Store", func() {
	var (
		mockCtrl *gomock.Controller
		mockMetadataStoreServiceClient *ml_metadata.MockMetadataStoreServiceClient
		store GrpcMetadataStore
		workflowName = randomString()
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

		When("GetArtifactsByContext does not return an artifact with the property", func() {
			It("filters out invalid artifacts", func() {
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{

							},
							{
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: randomString(),
										},
									},
									randomString(): {
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

		When("GetArtifactsByContext returns Artifacts with non-string properties of the same name", func() {
			It("filters out invalid artifacts", func() {
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: randomString(),
										},
									},
									PushedDestinationCustomProperty: {
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
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{

							},
							{
								CustomProperties: map[string]*ml_metadata.Value{
									randomString(): {
										Value: &ml_metadata.Value_StringValue{
											StringValue: randomString(),
										},
									},
									PushedDestinationCustomProperty: {
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

		When("GetArtifactsByContext returns artifacts with the property", func() {
			It("Returns all property values", func() {
				contextId := givenContextId()
				mockMetadataStoreServiceClient.EXPECT().
					GetArtifactsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByContextRequest{ContextId: &contextId})).
					Return(&ml_metadata.GetArtifactsByContextResponse{
						Artifacts: []*ml_metadata.Artifact{
							{
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "first-model",
										},
									},
									PushedDestinationCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "some://where",
										},
									},
								},
							},
							{
								CustomProperties: map[string]*ml_metadata.Value{
									NameCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "second-model",
										},
									},
									PushedDestinationCustomProperty: {
										Value: &ml_metadata.Value_StringValue{
											StringValue: "some://where.else",
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
						Name: "first-model",
						Location: "some://where",
					},
					ServingModelArtifact{
						Name: "second-model",
						Location: "some://where.else",
					},
				))
			})
		})
	})
})
