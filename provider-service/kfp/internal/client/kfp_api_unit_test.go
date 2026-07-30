//go:build unit

package client

import (
	"context"

	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
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

var _ = Context("createKfpApi", func() {
	newConfig := func(multiUser bool) config.Config {
		return config.Config{
			Parameters: config.Parameters{
				GrpcKfpApiAddress: "ml-pipeline:8887",
				KfpMultiUserMode:  multiUser,
			},
		}
	}

	// capturingDialer records the dial options it is given and returns a real
	// (lazy, unconnected) client so construction succeeds without a backend.
	var captured []grpc.DialOption
	capturingDialer := func(
		target string,
		opts ...grpc.DialOption,
	) (*grpc.ClientConn, error) {
		captured = opts
		return grpc.NewClient(target, opts...)
	}

	BeforeEach(func() {
		captured = nil
	})

	When("multi-user mode is enabled", func() {
		It("attaches the bearer credential dial option", func() {
			kfpApi, err := createKfpApi(
				context.Background(),
				newConfig(true),
				capturingDialer,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(kfpApi).ToNot(BeNil())
			// transport credentials + bearer credential
			Expect(captured).To(HaveLen(2))
		})
	})

	When("multi-user mode is disabled", func() {
		It("attaches no auth dial option", func() {
			kfpApi, err := createKfpApi(
				context.Background(),
				newConfig(false),
				capturingDialer,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(kfpApi).ToNot(BeNil())
			// transport credentials only
			Expect(captured).To(HaveLen(1))
		})
	})
})
