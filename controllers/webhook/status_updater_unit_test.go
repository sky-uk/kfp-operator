//go:build unit

package webhook

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Context("Handle", func() {
	var logger, _ = common.NewLogger(zapcore.DebugLevel)
	var ctx = logr.NewContext(context.Background(), logger)

	scheme := runtime.NewScheme()
	err := pipelineshub.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	var client client.Client
	var updater StatusUpdater

	Context("RunName is present in event", func() {
		var run pipelineshub.Run
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()
		rce.RunConfigurationName = nil

		BeforeEach(func() {
			rce.Status = common.RunCompletionStatuses.Succeeded

			run = pipelineshub.Run{}
			run.Status = pipelineshub.RunStatus{}
			run.Name = rce.RunName.Name
			run.Namespace = rce.RunName.Namespace

			client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelineshub.Run{}).
				Build()
			updater = StatusUpdater{client}
		})

		When("Run resource is found", func() {
			It("updates Run Status", func() {
				err = client.Create(context.Background(), &run)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, run.GetNamespacedName(), &run)
				Expect(err).ToNot(HaveOccurred())
				Expect(run.Status.CompletionState).
					To(Equal(pipelineshub.CompletionStates.Succeeded))
			})
		})

		When("Run resource is not found", func() {
			It("should return a MissingResourceError", func() {
				err = updater.Handle(ctx, rce)
				var expectedErr *MissingResourceError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
			})
		})

		When("event RunName has no Namespace", func() {
			It("should not return error and not update CompletionState", func() {
				err = client.Create(context.Background(), &run)
				Expect(err).ToNot(HaveOccurred())

				expectedState := run.Status.CompletionState
				rce.RunName.Namespace = ""
				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, run.GetNamespacedName(), &run)
				Expect(err).ToNot(HaveOccurred())
				Expect(run.Status.CompletionState).
					To(Equal(expectedState))
			})
		})

		When("k8s client operation fails", func() {
			It("should return error", func() {
				client = fake.NewClientBuilder().Build()
				updater = StatusUpdater{client}
				name := common.RandomNamespacedName()
				rce.RunName = &name
				err = updater.Handle(ctx, rce)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("RunConfigurationName is present in event", func() {
		var rc pipelineshub.RunConfiguration
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()
		rce.RunName = nil

		BeforeEach(func() {
			client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelineshub.RunConfiguration{}).
				Build()
			updater = StatusUpdater{client}

			rce.Status = common.RunCompletionStatuses.Succeeded

			rc = pipelineshub.RunConfiguration{}
			rc.Status = pipelineshub.RunConfigurationStatus{}
			rc.Name = rce.RunConfigurationName.Name
			rc.Namespace = rce.RunConfigurationName.Namespace
		})

		When("RunConfiguration resource is found", func() {
			It("updates the RunConfiguration ProviderId and Artifacts", func() {
				err = client.Create(context.Background(), &rc)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, rc.GetNamespacedName(), &rc)
				Expect(err).ToNot(HaveOccurred())
				Expect(rc.Status.LatestRuns.Succeeded.ProviderId).
					To(Equal(rce.RunId))
				Expect(rc.Status.LatestRuns.Succeeded.Artifacts).
					To(Equal(rce.Artifacts))
			})
		})

		When("RunConfiguration resource is not found", func() {
			It("should return not found error and not update the Status ProviderId and Artifacts", func() {
				expectedProviderId := rc.Status.LatestRuns.Succeeded.ProviderId
				expectedArtifacts := rc.Status.LatestRuns.Succeeded.Artifacts

				err = updater.Handle(ctx, rce)
				var expectedErr *MissingResourceError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())

				Expect(rc.Status.LatestRuns.Succeeded.ProviderId).
					To(Equal(expectedProviderId))

				Expect(rc.Status.LatestRuns.Succeeded.Artifacts).
					To(Equal(expectedArtifacts))
			})
		})

		When("event status is not Succeeded", func() {
			It("should not error and not update the Status ProviderId and Artifacts", func() {
				expectedProviderId := rc.Status.LatestRuns.Succeeded.ProviderId
				expectedArtifacts := rc.Status.LatestRuns.Succeeded.Artifacts

				rce.Status = common.RunCompletionStatuses.Failed
				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				Expect(rc.Status.LatestRuns.Succeeded.ProviderId).
					To(Equal(expectedProviderId))

				Expect(rc.Status.LatestRuns.Succeeded.Artifacts).
					To(Equal(expectedArtifacts))
			})
		})

		When("event RunConfiguration has no Namespace", func() {
			It("should not error and not update the Status ProviderId and Artifacts", func() {
				expectedProviderId := rc.Status.LatestRuns.Succeeded.ProviderId
				expectedArtifacts := rc.Status.LatestRuns.Succeeded.Artifacts

				rce.RunConfigurationName.Namespace = ""
				rce.Status = common.RunCompletionStatuses.Failed
				err = client.Create(context.Background(), &rc)

				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				Expect(rc.Status.LatestRuns.Succeeded.ProviderId).
					To(Equal(expectedProviderId))

				Expect(rc.Status.LatestRuns.Succeeded.Artifacts).
					To(Equal(expectedArtifacts))
			})
		})

		When("k8s client operation fails", func() {
			It("should return error", func() {
				client = fake.NewClientBuilder().Build()
				updater = StatusUpdater{client}
				name := common.RandomNamespacedName()
				rce.RunConfigurationName = &name
				err = updater.Handle(ctx, rce)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Both RunName and RunConfigurationName are present in event", func() {
		var rc pipelineshub.RunConfiguration
		var run pipelineshub.Run
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()

		BeforeEach(func() {
			run = pipelineshub.Run{}
			run.Status = pipelineshub.RunStatus{}
			run.Name = rce.RunName.Name
			run.Namespace = rce.RunName.Namespace

			rc = pipelineshub.RunConfiguration{}
			rc.Status = pipelineshub.RunConfigurationStatus{}
			rc.Name = rce.RunConfigurationName.Name
			rc.Namespace = rce.RunConfigurationName.Namespace

			rce.Status = common.RunCompletionStatuses.Succeeded
			rce.RunName = &common.NamespacedName{Name: run.Name, Namespace: run.Namespace}

			client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelineshub.Run{}, &pipelineshub.RunConfiguration{}).
				Build()
			updater = StatusUpdater{client}
		})

		When("RunConfiguration resource is found", func() {
			It("updates the RunConfiguration ProviderId and Artifacts", func() {
				err = client.Create(context.Background(), &rc)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, rc.GetNamespacedName(), &rc)
				Expect(err).ToNot(HaveOccurred())
				Expect(rc.Status.LatestRuns.Succeeded.ProviderId).
					To(Equal(rce.RunId))
				Expect(rc.Status.LatestRuns.Succeeded.Artifacts).
					To(Equal(rce.Artifacts))
			})
		})

		When("Run resource is found", func() {
			It("updates the Run", func() {
				err = client.Create(context.Background(), &run)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(ctx, rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, run.GetNamespacedName(), &run)
				Expect(err).ToNot(HaveOccurred())
				Expect(run.Status.CompletionState).
					To(Equal(pipelineshub.CompletionStates.Succeeded))
			})
		})

		When("No resource is found", func() {
			It("returns an error", func() {
				err = updater.Handle(ctx, rce)
				var expectedErr *MissingResourceError
				Expect(errors.As(err, &expectedErr)).To(BeTrue())
			})
		})
	})
})
