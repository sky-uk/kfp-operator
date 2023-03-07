//go:build unit
// +build unit

package kfp

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
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
		runId = common.RandomString()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("GetResourceReferences", func() {
		When("GetRun errors", func() {
			It("Errors", func() {
				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(nil, fmt.Errorf("an error"))

				_, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(HaveOccurred())
			})
		})

		When("GetRun returns run with a JOB but without CREATOR", func() {
			It("Returns empty ResourceReferences", func() {
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Name:         common.RandomString(),
								Relationship: go_client.Relationship_UNKNOWN_RELATIONSHIP,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   common.RandomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&mockRunDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences).To(Equal(ResourceReferences{}))
			})
		})

		When("GetRun returns run with a CREATOR that is not a JOB", func() {
			It("Returns empty ResourceReferences", func() {
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Name:         common.RandomString(),
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_PIPELINE,
									Id:   common.RandomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&mockRunDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences).To(Equal(ResourceReferences{}))
			})
		})

		When("GetRun returns run with JOB as CREATOR", func() {
			It("Returns populated ResourceReferences", func() {
				runConfigurationName := common.RandomString()
				runName := common.RandomNamespacedName()

				runDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id:   runId,
						Name: runName.Name,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Name:         runConfigurationName,
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   common.RandomString(),
								},
							},
							{
								Relationship: go_client.Relationship_OWNER,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_NAMESPACE,
									Id:   runName.Namespace,
								},
							},
						},
					},
				}

				mockRunServiceClient.EXPECT().
					GetRun(gomock.Any(), gomock.Eq(&go_client.GetRunRequest{RunId: runId})).
					Return(&runDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences.RunConfigurationName).To(Equal(runConfigurationName))
				Expect(resourceReferences.RunName).To(Equal(runName))
			})
		})
	})
})
