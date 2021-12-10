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

	When("The service finds the artifact", func() {
		It("Returns a ModelArtifact", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: nil})).
				Return(&ml_metadata.GetArtifactsByIDResponse{
				Artifacts: []*ml_metadata.Artifact{
					{

					},
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
			result, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(ModelArtifact{
				PushDestination: "some://where",
			}))
		})
	})

	When("the response errors", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: nil})).
				Return(nil, fmt.Errorf("an error"))

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the response does not contain an artifact with a push destination", func() {
		It("Errors", func() {
			mockMetadataStoreServiceClient.EXPECT().
				GetArtifactsByID(gomock.Any(), gomock.Eq(&ml_metadata.GetArtifactsByIDRequest{ArtifactIds: nil})).
				Return(&ml_metadata.GetArtifactsByIDResponse{}, nil)

			_, err := store.GetServingModelArtifact(context.Background(), pipelineName, workflowName)
			Expect(err).To(HaveOccurred())
		})
	})
})
