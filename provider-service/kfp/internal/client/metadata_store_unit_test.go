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
})
