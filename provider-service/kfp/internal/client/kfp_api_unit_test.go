//go:build unit

package client

import (
	"context"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Context("KFP API", func() {
	var (
		mockRunServiceClient          mocks.MockRunServiceClient
		mockRecurringRunServiceClient mocks.MockRecurringRunServiceClient
		kfpApi                        GrpcKfpApi
		runId                         string
	)

	BeforeEach(func() {
		mockRunServiceClient = mocks.MockRunServiceClient{}
		mockRecurringRunServiceClient = mocks.MockRecurringRunServiceClient{}
		kfpApi = GrpcKfpApi{
			RunServiceClient:          &mockRunServiceClient,
			RecurringRunServiceClient: &mockRecurringRunServiceClient,
		}
		runId = common.RandomString()
	})

	Describe("GetResourceReferences", func() {
		When("GetRun returns run with a job with parameter label fields set", func() {
			It("Returns a populated ResourceReference", func() {
				mockRunDetail := go_client.Run{
					RuntimeConfig: &go_client.RuntimeConfig{
						Parameters: map[string]*structpb.Value{
							label.RunName:                   structpb.NewStringValue("RunName"),
							label.RunNamespace:              structpb.NewStringValue("RunNamespace"),
							label.RunConfigurationName:      structpb.NewStringValue("RunConfigurationName"),
							label.RunConfigurationNamespace: structpb.NewStringValue("RunConfigurationNamespace"),
							label.PipelineName:              structpb.NewStringValue("PipelineName"),
							label.PipelineNamespace:         structpb.NewStringValue("PipelineNamespace"),
						},
					},
				}

				mockRunServiceClient.On(
					"GetRun",
					&go_client.GetRunRequest{RunId: runId},
				).Return(&mockRunDetail, nil)

				resourceReferences, err := kfpApi.GetResourceReferences(context.Background(), runId)
				Expect(err).To(BeNil())
				Expect(resourceReferences).To(Equal(resource.References{
					PipelineName: common.NamespacedName{
						Name:      "PipelineName",
						Namespace: "PipelineNamespace",
					},
					RunConfigurationName: common.NamespacedName{
						Name:      "RunConfigurationName",
						Namespace: "RunConfigurationNamespace",
					},
					RunName: common.NamespacedName{
						Name:      "RunName",
						Namespace: "RunNamespace",
					},
				}))
			})
		})

		When("GetRun returns run with missing parameter fields", func() {
			It("Returns empty ResourceReferences", func() {
				mockRunDetail := go_client.Run{
					RuntimeConfig: &go_client.RuntimeConfig{
						Parameters: map[string]*structpb.Value{},
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
	})
})
