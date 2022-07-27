//go:build unit
// +build unit

package run_completion

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("KFP API", func() {
	var (
		mockCtrl             *gomock.Controller
		mockRunServiceClient *MockRunServiceClient
		kfpApi               GrpcKfpApi
		runId                string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockRunServiceClient = NewMockRunServiceClient(mockCtrl)
		kfpApi = GrpcKfpApi{RunServiceClient: mockRunServiceClient}
		runId = randomString()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("GetRunConfiguration", func() {
		When("GetRun errors", func() {
			It("Errors", func() {
				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(nil, fmt.Errorf("an error"))

				_, err := kfpApi.GetRunConfiguration(context.Background(), runId)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetRun returns run with JOB but without CREATOR", func() {
			It("Returns empty string", func() {
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							&go_client.ResourceReference{
								Name:         randomString(),
								Relationship: go_client.Relationship_UNKNOWN_RELATIONSHIP,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   randomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&mockRunDetail, nil)

				runConfig, err := kfpApi.GetRunConfiguration(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(runConfig).To(BeEmpty())
			})
		})

		When("GetRun returns run without JOB", func() {
			It("Returns empty string", func() {
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							&go_client.ResourceReference{
								Name:         randomString(),
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_PIPELINE,
									Id:   randomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&mockRunDetail, nil)

				runConfig, err := kfpApi.GetRunConfiguration(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(runConfig).To(BeEmpty())
			})
		})

		When("GetRun returns run with JOB as CREATOR", func() {
			It("Returns RunConfiguration name", func() {
				mockRunConfig := randomString()
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							&go_client.ResourceReference{
								Name:         mockRunConfig,
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   randomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&mockRunDetail, nil)

				runConfig, err := kfpApi.GetRunConfiguration(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(runConfig).To(Equal(mockRunConfig))
			})
		})
	})
})
