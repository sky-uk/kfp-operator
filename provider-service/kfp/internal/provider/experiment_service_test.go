//go:build unit

package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/util"
)

var _ = Describe("ExperimentService", func() {
	var (
		mockClient        mocks.MockExperimentServiceClient
		experimentService ExperimentService
		nsn               = common.NamespacedName{
			Name:      "name",
			Namespace: "namespace",
		}
		ctx = context.Background()
	)

	BeforeEach(func() {
		mockClient = mocks.MockExperimentServiceClient{}
		experimentService = &DefaultExperimentService{
			&mockClient,
		}
	})

	Context("CreateExperiment", func() {
		It("should return experiment result id ", func() {
			expectedId := "expected-result-id"
			mockClient.On(
				"CreateExperiment",
				&go_client.CreateExperimentRequest{
					Experiment: &go_client.Experiment{
						DisplayName: "namespace-name",
						Description: "description",
					},
				},
			).Return(&go_client.Experiment{ExperimentId: expectedId}, nil)
			res, err := experimentService.CreateExperiment(ctx, nsn, "description")

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(expectedId))
		})

		When("experiment namespaced name is invalid", func() {
			It("should return error", func() {
				invalidNsn := common.NamespacedName{
					Namespace: "namespace",
				}
				res, err := experimentService.CreateExperiment(ctx, invalidNsn, "foo")

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient CreateExperiment errors", func() {
			It("should return error", func() {
				mockClient.On(
					"CreateExperiment",
					&go_client.CreateExperimentRequest{
						Experiment: &go_client.Experiment{
							DisplayName: "namespace-name",
							Description: "description",
						},
					},
				).Return(nil, errors.New("failed"))
				res, err := experimentService.CreateExperiment(ctx, nsn, "description")

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})

	Context("DeleteExperiment", func() {
		It("should not error", func() {
			id := "delete-experiment-id"
			mockClient.On(
				"DeleteExperiment",
				&go_client.DeleteExperimentRequest{
					ExperimentId: id,
				},
			).Return(nil)
			err := experimentService.DeleteExperiment(ctx, id)

			Expect(err).ToNot(HaveOccurred())
		})

		When("experimentServiceClient DeleteExperiment errors", func() {
			It("should return error", func() {
				id := "delete-experiment-id"
				mockClient.On(
					"DeleteExperiment",
					&go_client.DeleteExperimentRequest{
						ExperimentId: id,
					},
				).Return(errors.New("failed"))
				err := experimentService.DeleteExperiment(ctx, id)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("ExperimentIdByDisplayName", func() {
		It("should return the experiment result id", func() {
			expectedResult := go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{
					{ExperimentId: "one"},
				},
			}
			mockClient.On(
				"ListExperiments",
				&go_client.ListExperimentsRequest{
					Filter: util.ByDisplayNameFilter("namespace-name"),
				},
			).Return(&expectedResult, nil)
			res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})

		When("experiment namespaced name is invalid", func() {
			It("should return error", func() {
				invalidNsn := common.NamespacedName{
					Namespace: "namespace",
				}
				res, err := experimentService.ExperimentIdByDisplayName(ctx, invalidNsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment errors", func() {
			It("should return error", func() {
				mockClient.On(
					"ListExperiments",
					&go_client.ListExperimentsRequest{
						Filter: util.ByDisplayNameFilter("namespace-name"),
					},
				).Return(nil, errors.New("failed"))
				res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment returns no experiments", func() {
			It("should return error", func() {
				expectedResult := go_client.ListExperimentsResponse{
					Experiments: []*go_client.Experiment{},
				}
				mockClient.On(
					"ListExperiments",
					&go_client.ListExperimentsRequest{
						Filter: util.ByDisplayNameFilter("namespace-name"),
					},
				).Return(&expectedResult, nil)
				res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment returns > 1 experiments", func() {
			It("should return error", func() {
				expectedResult := go_client.ListExperimentsResponse{
					Experiments: []*go_client.Experiment{
						{ExperimentId: "one"},
						{ExperimentId: "two"},
					},
				}
				mockClient.On(
					"ListExperiments",
					&go_client.ListExperimentsRequest{
						Filter: util.ByDisplayNameFilter("namespace-name"),
					},
				).Return(&expectedResult, nil)
				res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})
})
