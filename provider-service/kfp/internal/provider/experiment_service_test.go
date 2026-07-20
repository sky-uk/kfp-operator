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
	"github.com/stretchr/testify/mock"
)

var _ = Describe("ExperimentService", func() {
	const providerNamespace = "kfp"

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
	})

	// namespaceScopedSpecs asserts the request namespace the service sends on the
	// operations that carry one. expectedNamespace is the namespace the service is
	// expected to send, independent of the resource's own namespace.
	namespaceScopedSpecs := func(expectedNamespace string) {
		It("CreateExperiment sends the expected namespace", func() {
			expectedId := "expected-result-id"
			mockClient.On(
				"CreateExperiment",
				&go_client.CreateExperimentRequest{
					Experiment: &go_client.Experiment{
						DisplayName: "namespace-name",
						Description: "description",
						Namespace:   expectedNamespace,
					},
				},
			).Return(&go_client.Experiment{ExperimentId: expectedId}, nil)
			res, err := experimentService.CreateExperiment(ctx, nsn, "description")

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(expectedId))
		})

		It("ExperimentIdByDisplayName sends the expected namespace", func() {
			expectedResult := go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{
					{ExperimentId: "one"},
				},
			}
			mockClient.On(
				"ListExperiments",
				&go_client.ListExperimentsRequest{
					Filter:    util.ByDisplayNameFilter("namespace-name"),
					Namespace: expectedNamespace,
				},
			).Return(&expectedResult, nil)
			res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})
	}

	Context("single-user mode", func() {
		BeforeEach(func() {
			experimentService = &DefaultExperimentService{
				client:           &mockClient,
				requestNamespace: "",
			}
		})

		namespaceScopedSpecs("")

		It("sends an empty namespace even though the resource is namespaced", func() {
			mockClient.On(
				"CreateExperiment",
				mock.MatchedBy(func(req *go_client.CreateExperimentRequest) bool {
					return req.Experiment.Namespace == "" &&
						req.Experiment.DisplayName == "namespace-name"
				}),
			).Return(&go_client.Experiment{ExperimentId: "id"}, nil)
			_, err := experimentService.CreateExperiment(ctx, nsn, "description")

			Expect(err).ToNot(HaveOccurred())
		})

		It("leaves an unqualified experiment name unscoped", func() {
			unqualified := common.NamespacedName{Name: "Default"}
			mockClient.On(
				"ListExperiments",
				&go_client.ListExperimentsRequest{
					Filter:    util.ByDisplayNameFilter("Default"),
					Namespace: "",
				},
			).Return(&go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{{ExperimentId: "one"}},
			}, nil)
			res, err := experimentService.ExperimentIdByDisplayName(ctx, unqualified)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})
	})

	Context("multi-user mode", func() {
		BeforeEach(func() {
			experimentService = &DefaultExperimentService{
				client:           &mockClient,
				requestNamespace: providerNamespace,
			}
		})

		namespaceScopedSpecs(providerNamespace)

		It("sends the provider namespace, not the resource namespace", func() {
			mockClient.On(
				"ListExperiments",
				mock.MatchedBy(func(req *go_client.ListExperimentsRequest) bool {
					return req.Namespace == providerNamespace &&
						req.Namespace != nsn.Namespace
				}),
			).Return(&go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{{ExperimentId: "one"}},
			}, nil)
			res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})

		It("scopes an unqualified experiment name to the provider namespace", func() {
			unqualified := common.NamespacedName{Name: "default"}
			mockClient.On(
				"ListExperiments",
				&go_client.ListExperimentsRequest{
					Filter:    util.ByDisplayNameFilter("kfp-default"),
					Namespace: providerNamespace,
				},
			).Return(&go_client.ListExperimentsResponse{
				Experiments: []*go_client.Experiment{{ExperimentId: "one"}},
			}, nil)
			res, err := experimentService.ExperimentIdByDisplayName(ctx, unqualified)

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("one"))
		})
	})

	Context("mode-independent", func() {
		BeforeEach(func() {
			experimentService = &DefaultExperimentService{
				client:           &mockClient,
				requestNamespace: providerNamespace,
			}
		})

		Context("CreateExperiment", func() {
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
								Namespace:   providerNamespace,
							},
						},
					).Return(nil, errors.New("failed"))
					res, err := experimentService.CreateExperiment(ctx, nsn, "description")

					Expect(err).To(HaveOccurred())
					Expect(res).To(BeEmpty())
				})
			})
		})

		Context("ExperimentIdByDisplayName", func() {
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
							Filter:    util.ByDisplayNameFilter("namespace-name"),
							Namespace: providerNamespace,
						},
					).Return(nil, errors.New("failed"))
					res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

					Expect(err).To(HaveOccurred())
					Expect(res).To(BeEmpty())
				})
			})

			When("experimentServiceClient ListExperiment returns no experiments", func() {
				It("should return error", func() {
					mockClient.On(
						"ListExperiments",
						&go_client.ListExperimentsRequest{
							Filter:    util.ByDisplayNameFilter("namespace-name"),
							Namespace: providerNamespace,
						},
					).Return(&go_client.ListExperimentsResponse{
						Experiments: []*go_client.Experiment{},
					}, nil)
					res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

					Expect(err).To(HaveOccurred())
					Expect(res).To(BeEmpty())
				})
			})

			When("experimentServiceClient ListExperiment returns > 1 experiments", func() {
				It("should return error", func() {
					mockClient.On(
						"ListExperiments",
						&go_client.ListExperimentsRequest{
							Filter:    util.ByDisplayNameFilter("namespace-name"),
							Namespace: providerNamespace,
						},
					).Return(&go_client.ListExperimentsResponse{
						Experiments: []*go_client.Experiment{
							{ExperimentId: "one"},
							{ExperimentId: "two"},
						},
					}, nil)
					res, err := experimentService.ExperimentIdByDisplayName(ctx, nsn)

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

	})
})
