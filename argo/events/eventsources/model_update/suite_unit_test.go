//go:build unit
// +build unit

package main

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/util/rand"
	"pipelines.kubeflow.org/events/ml_metadata"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Update Unit Suite")
}

var _ = Describe("gRPC Metadata Store", func() {
	var (
		mockCtrl *gomock.Controller
		mockMetadataStoreServiceClient *ml_metadata.MockMetadataStoreServiceClient
		store GrpcMetadataStore
		pipelineName = rand.String(5)
		workflowName = rand.String(5)
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

	var givenExecutionIDs = func() []int64 {
		executionIDs := make([]int64, rand.IntnRange(2, 10))
		executions := make([]*ml_metadata.Execution, len(executionIDs))

		for i, _ := range executionIDs {
			executionIDs[i] = int64(rand.Int())
			executions[i] = &ml_metadata.Execution{
				Id: &executionIDs[i],
			}
		}
		mockMetadataStoreServiceClient.EXPECT().
			GetExecutionsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetExecutionsByContextRequest{ContextId: nil})).
			Return(&ml_metadata.GetExecutionsByContextResponse{
				Executions: executions,
			}, nil)

		return executionIDs
	}

	var givenArtifactIDs = func() []int64 {
		artifactIDs := make([]int64, rand.IntnRange(2, 10))
		events := make([]*ml_metadata.Event, len(artifactIDs))

		for i, _ := range artifactIDs {
			artifactIDs[i] = int64(rand.Int())
			events[i] = &ml_metadata.Event{
				ArtifactId: &artifactIDs[i],
			}
		}
		mockMetadataStoreServiceClient.EXPECT().
			GetEventsByExecutionIDs(gomock.Any(), gomock.Eq(&ml_metadata.GetEventsByExecutionIDsRequest{ExecutionIds: givenExecutionIDs()})).
			Return(&ml_metadata.GetEventsByExecutionIDsResponse{
				Events: events,
			}, nil)

		return artifactIDs
	}

	When("GetExecutionsByContext errors", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetExecutionsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetExecutionsByContextRequest{ContextId: nil})).
				Return(nil, fmt.Errorf("an error"))

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("GetExecutionsByContext returns no executions", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetExecutionsByContext(gomock.Any(), gomock.Eq(&ml_metadata.GetExecutionsByContextRequest{ContextId: nil})).
				Return(&ml_metadata.GetExecutionsByContextResponse{
					Executions: nil,
			}, nil)

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("GetEventsByExecutionIDs errors", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetEventsByExecutionIDs(gomock.Any(), gomock.Eq(&ml_metadata.GetEventsByExecutionIDsRequest{ExecutionIds: givenExecutionIDs()})).
				Return(nil, fmt.Errorf("an error"))

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("GetEventsByExecutionIDs returns no events", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetEventsByExecutionIDs(gomock.Any(), gomock.Eq(&ml_metadata.GetEventsByExecutionIDsRequest{ExecutionIds: givenExecutionIDs()})).
				Return(&ml_metadata.GetEventsByExecutionIDsResponse{
					Events: nil,
				}, nil)

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("GetEventsByExecutionIDs returns an event", func() {
		It("Filters out missing artifact ids", func() {
			artifactId := int64(rand.Int())

			mockMetadataStoreServiceClient.EXPECT().
				GetEventsByExecutionIDs(gomock.Any(), gomock.Eq(&ml_metadata.GetEventsByExecutionIDsRequest{ExecutionIds: givenExecutionIDs()})).
				Return(&ml_metadata.GetEventsByExecutionIDsResponse{
					Events: []*ml_metadata.Event{
						{
							ArtifactId: nil,
						},
						{
							ArtifactId: &artifactId,
						},
					},
				}, nil)
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: []int64{artifactId}})).
				Return(&ml_metadata.GetArtifactsByIDResponse{}, nil)

			store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
		})
	})

	When("GetArtifactsByIDRequest errors", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: givenArtifactIDs()})).
				Return(nil, fmt.Errorf("an error"))

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("GetArtifactsByIDRequest does not return an artifact with a push destination", func() {
		It("returns an empty list", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: givenArtifactIDs()})).
				Return(&ml_metadata.GetArtifactsByIDResponse{}, nil)

			results, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(BeEmpty())
		})
	})

	When("GetArtifactsByIDRequest returns artifacts with a push destination", func() {
		It("Returns all ModelArtifacts", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: givenArtifactIDs()})).
				Return(&ml_metadata.GetArtifactsByIDResponse{
				Artifacts: []*ml_metadata.Artifact{
					{
						CustomProperties: map[string]*ml_metadata.Value{
							PushedDestinationCustomProperty: {
								Value: &ml_metadata.Value_IntValue{
									IntValue: 42,
								},
							},
						},
					},
					{
						CustomProperties: map[string]*ml_metadata.Value{
							PushedDestinationCustomProperty: {
								Value: &ml_metadata.Value_StringValue{
									StringValue: "some://where",
								},
							},
						},
					},
					{
						CustomProperties: map[string]*ml_metadata.Value{
							PushedDestinationCustomProperty: {
								Value: &ml_metadata.Value_StringValue{
									StringValue: "some://where.else",
								},
							},
						},
					},
				},
			}, nil)
			results, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).NotTo(HaveOccurred())

			Expect(results).To(ContainElements(ModelArtifact{
				PushDestination: "some://where",
			}, ModelArtifact{
				PushDestination: "some://where.else",
			}))
		})
	})
})
