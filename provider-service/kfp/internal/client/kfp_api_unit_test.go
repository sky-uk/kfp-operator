//go:build unit

package client

import (
	"context"

	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Context("KFP API", func() {
	var (
		mockRunServiceClient mocks.MockRunServiceClient
		kfpApi               GrpcKfpApi
		runId                string
	)

	BeforeEach(func() {
		mockRunServiceClient = mocks.MockRunServiceClient{}
		kfpApi = GrpcKfpApi{
			RunServiceClient: &mockRunServiceClient,
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
				Expect(err).NotTo(HaveOccurred())
				Expect(resourceReferences).To(BeComparableTo(resource.References{
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
				}, cmpopts.IgnoreFields(resource.References{}, "FinishedAt")))
			})
		})
	})
})

var _ = Context("GrpcDialOptions", func() {
	When("a bearer token source is provided", func() {
		It("attaches the bearer credential dial option", func() {
			opts := GrpcDialOptions(
				oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token"}),
			)
			// transport credentials + bearer credential
			Expect(opts).To(HaveLen(2))
		})
	})

	When("no bearer token source is provided", func() {
		It("attaches no auth dial option", func() {
			opts := GrpcDialOptions(nil)
			// transport credentials only
			Expect(opts).To(HaveLen(1))
		})
	})
})
