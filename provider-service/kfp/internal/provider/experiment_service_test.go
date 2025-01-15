//go:build unit

package provider

import (
	"context"
	"errors"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/mocks"
)

var _ = Describe("ExperimentService", func() {
	var (
		mockExperimentServiceClient mocks.MockExperimentServiceClient
		experimentService           ExperimentService
		nsn                         = common.NamespacedName{
			Name:      "name",
			Namespace: "namespace",
		}
	)

	BeforeEach(func() {
		mockExperimentServiceClient = mocks.MockExperimentServiceClient{}
		experimentService = &DefaultExperimentService{
			context.Background(),
			&mockExperimentServiceClient,
		}
	})

	Context("ExperimentIdByName", func() {
		It("should return the experiment result id", func() {
			expectedResult := go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{
					{Id: "one"},
				},
			}
			mockExperimentServiceClient.On(
				"ListExperiment",
				&go_client.ListExperimentsRequest{
					Filter: *byNameFilter("namespace-name"),
				},
			).Return(&expectedResult, nil)
			res, err := experimentService.ExperimentIdByName(nsn)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})

		When("experiment namespaced name is invalid", func() {
			It("should return error", func() {
				invalidNsn := common.NamespacedName{
					Namespace: "namespace",
				}
				res, err := experimentService.ExperimentIdByName(invalidNsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment errors", func() {
			It("should return error", func() {
				mockExperimentServiceClient.On(
					"ListExperiment",
					&go_client.ListExperimentsRequest{
						Filter: *byNameFilter("namespace-name"),
					},
				).Return(nil, errors.New("failed"))
				res, err := experimentService.ExperimentIdByName(nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment returns no experiments", func() {
			It("should return error", func() {
				expectedResult := go_client.ListExperimentsResponse{
					Experiments: []*go_client.Experiment{},
				}
				mockExperimentServiceClient.On(
					"ListExperiment",
					&go_client.ListExperimentsRequest{
						Filter: *byNameFilter("namespace-name"),
					},
				).Return(&expectedResult, nil)
				res, err := experimentService.ExperimentIdByName(nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})

		When("experimentServiceClient ListExperiment returns > 1 experiments", func() {
			It("should return error", func() {
				expectedResult := go_client.ListExperimentsResponse{
					Experiments: []*go_client.Experiment{
						{Id: "one"},
						{Id: "two"},
					},
				}
				mockExperimentServiceClient.On(
					"ListExperiment",
					&go_client.ListExperimentsRequest{
						Filter: *byNameFilter("namespace-name"),
					},
				).Return(&expectedResult, nil)
				res, err := experimentService.ExperimentIdByName(nsn)

				Expect(err).To(HaveOccurred())
				Expect(res).To(BeEmpty())
			})
		})
	})
})
