//go:build unit

package client

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common"
	"github.com/sky-uk/kfp-operator/common/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Context("KFP API", func() {
	var (
		mockRunServiceClient mocks.MockRunServiceClient
		mockJobServiceClient mocks.MockJobServiceClient
		kfpApi               GrpcKfpApi
		runId                string
	)

	BeforeEach(func() {
		mockRunServiceClient = mocks.MockRunServiceClient{}
		mockJobServiceClient = mocks.MockJobServiceClient{}
		kfpApi = GrpcKfpApi{
			RunServiceClient: &mockRunServiceClient,
			JobServiceClient: &mockJobServiceClient,
		}
		runId = testutil.RandomString()
	})

	Describe("GetResourceReferences", func() {
		When("GetRun errors", func() {
			It("Errors", func() {
				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(nil, errors.New("failed"))

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
								Name:         testutil.RandomString(),
								Relationship: go_client.Relationship_UNKNOWN_RELATIONSHIP,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   testutil.RandomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(&mockRunDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences).To(Equal(resource.References{}))
			})
		})

		When("GetRun returns run with a CREATOR that is not a JOB", func() {
			It("Returns empty ResourceReferences", func() {
				mockRunDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id: runId,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Name:         testutil.RandomString(),
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_PIPELINE,
									Id:   testutil.RandomString(),
								},
							},
						},
					},
				}

				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(&mockRunDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences).To(Equal(resource.References{}))
			})
		})

		When("GetRun returns run with JOB without Description as CREATOR", func() {
			It("Returns populated ResourceReferences", func() {
				runConfigurationName := common.NamespacedName{
					Name: testutil.RandomString(),
				}
				runName := testutil.RandomNamespacedName()

				jobId := testutil.RandomString()

				runDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id:   runId,
						Name: runName.Name,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Name:         runConfigurationName.Name,
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   jobId,
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

				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(&runDetail, nil)

				mockJobServiceClient.On(
					"GetJob",
					&go_client.GetJobRequest{Id: jobId},
				).Return(&go_client.Job{}, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences.RunConfigurationName).To(Equal(runConfigurationName))
				Expect(resourceReferences.RunName).To(Equal(runName))
			})
		})

		When("GetRun returns run with JOB with Description as CREATOR", func() {
			It("Returns populated ResourceReferences", func() {
				runConfigurationName := testutil.RandomNamespacedName()
				runName := common.NamespacedName{
					Name: testutil.RandomString(),
				}
				jobId := testutil.RandomString()

				runDetail := go_client.RunDetail{
					Run: &go_client.Run{
						Id:   runId,
						Name: runName.Name,
						ResourceReferences: []*go_client.ResourceReference{
							{
								Relationship: go_client.Relationship_CREATOR,
								Key: &go_client.ResourceKey{
									Type: go_client.ResourceType_JOB,
									Id:   jobId,
								},
							},
						},
					},
				}

				jobDetail := go_client.Job{
					Description: "runConfigurationName: " + testutil.UnsafeValue(runConfigurationName.String()),
				}

				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(&runDetail, nil)

				mockJobServiceClient.On(
					"GetJob",
					&go_client.GetJobRequest{Id: jobId},
				).Return(&jobDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences.RunConfigurationName).To(Equal(runConfigurationName))
			})
		})
	})
})
