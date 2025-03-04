//go:build unit

package webhook

import (
	"context"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
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
	err := pipelinesv1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	var client client.Client
	var updater StatusUpdater

	Context("event RunName is present", func() {
		var run pipelinesv1.Run
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()

		BeforeEach(func() {
			rce.Status = common.RunCompletionStatuses.Succeeded

			run = pipelinesv1.Run{}
			run.Status = pipelinesv1.RunStatus{}
			run.Name = rce.RunName.Name
			run.Namespace = rce.RunName.Namespace

			client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelinesv1.Run{}).
				Build()
			updater = StatusUpdater{ctx, client}
		})

		When("Run resource is found", func() {
			It("updates Run Status", func() {
				err = client.Create(context.Background(), &run)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(rce)
				Expect(err).ToNot(HaveOccurred())

				err = client.Get(ctx, run.GetNamespacedName(), &run)
				Expect(err).ToNot(HaveOccurred())
				Expect(run.Status.CompletionState).
					To(Equal(pipelinesv1.CompletionStates.Succeeded))
			})
		})

		When("Run resource is not found", func() {
			It("should not return error", func() {
				err = updater.Handle(rce)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("event RunName has no Namespace", func() {
			It("should not return error and not update CompletionState", func() {
				err = client.Create(context.Background(), &run)
				Expect(err).ToNot(HaveOccurred())

				expectedState := run.Status.CompletionState
				rce.RunName.Namespace = ""
				err = updater.Handle(rce)
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
				updater = StatusUpdater{ctx, client}
				err = updater.Handle(rce)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("event RunConfigurationName is present", func() {
		var rc pipelinesv1.RunConfiguration
		rce := RandomRunCompletionEventData().ToRunCompletionEvent()

		BeforeEach(func() {
			client = fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelinesv1.RunConfiguration{}).
				Build()
			updater = StatusUpdater{ctx, client}

			rce.Status = common.RunCompletionStatuses.Succeeded

			rc = pipelinesv1.RunConfiguration{}
			rc.Status = pipelinesv1.RunConfigurationStatus{}
			rc.Name = rce.RunConfigurationName.Name
			rc.Namespace = rce.RunConfigurationName.Namespace
		})

		When("RunConfiguration resource is found", func() {
			It("updates the RunConfiguration ProviderId and Artifacts", func() {
				err = client.Create(context.Background(), &rc)
				Expect(err).ToNot(HaveOccurred())

				err = updater.Handle(rce)
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
			It("should not error and not update the Status ProviderId and Artifacts", func() {
				expectedProviderId := rc.Status.LatestRuns.Succeeded.ProviderId
				expectedArtifacts := rc.Status.LatestRuns.Succeeded.Artifacts

				err = updater.Handle(rce)
				Expect(err).ToNot(HaveOccurred())

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
				err = updater.Handle(rce)
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

				err = updater.Handle(rce)
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
				updater = StatusUpdater{ctx, client}
				err = updater.Handle(rce)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
