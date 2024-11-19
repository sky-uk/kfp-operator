//go:build decoupled

package status_updater

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient     client.Client
	statusUpdater StatusUpdater
	testEnv       *envtest.Environment
)

func TestStatusUpdaterDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Updater Decoupled Suite")
}

var _ = BeforeSuite(func() {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = pipelinesv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	statusUpdater = StatusUpdater{
		K8sClient: k8sClient,
	}
})

var _ = AfterSuite(func() {
	testEnv.Stop()
})

var _ = Describe("Status Updater", Serial, func() {
	Context("Runs", func() {
		var CompletionStateHasChangedTo = func(expectedState pipelinesv1.CompletionState) func(pipelinesv1.Run, pipelinesv1.Run) {
			return func(oldRun pipelinesv1.Run, newRun pipelinesv1.Run) {
				Expect(oldRun.Status.CompletionState).NotTo(Equal(expectedState))
				Expect(newRun.Status.CompletionState).To(Equal(expectedState))
			}
		}

		var HasNotChanged = func() func(pipelinesv1.Run, pipelinesv1.Run) {
			return func(oldRun pipelinesv1.Run, newRun pipelinesv1.Run) {
				Expect(newRun.GetStatus()).To(Equal(oldRun.GetStatus()))
			}
		}

		DescribeTable("updates Run on known states only",
			func(status common.RunCompletionStatus, expectation func(pipelinesv1.Run, pipelinesv1.Run)) {
				ctx := context.Background()
				run := pipelinesv1.RandomRun(apis.RandomLowercaseString())
				Expect(k8sClient.Create(ctx, run)).To(Succeed())

				runCompletionEvent := common.RunCompletionEvent{Status: status, RunName: &common.NamespacedName{
					Name:      run.Name,
					Namespace: run.Namespace,
				}}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())

				fetchedRun := pipelinesv1.Run{}
				Expect(k8sClient.Get(ctx, run.GetNamespacedName(), &fetchedRun)).To(Succeed())
				expectation(*run, fetchedRun)
			},
			Entry("succeeded should succeed", common.RunCompletionStatuses.Succeeded, CompletionStateHasChangedTo(pipelinesv1.CompletionStates.Succeeded)),
			Entry("failed should fail", common.RunCompletionStatuses.Failed, CompletionStateHasChangedTo(pipelinesv1.CompletionStates.Failed)),
			Entry("unknown should not override", common.RunCompletionStatus(""), HasNotChanged()))

		When("the run is not found", func() {
			It("do nothing", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
					Name:      common.RandomString(),
					Namespace: common.RandomString(),
				}}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())
			})
		})

		When("the runName has no namespace", func() {
			It("do nothing", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
					Name: common.RandomString(),
				}}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())
			})
		})

		When("the k8s API is unreachable", func() {
			It("errors", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
					Name:      common.RandomString(),
					Namespace: common.RandomString(),
				}}

				Expect((&StatusUpdater{
					NewFailingClient(),
				}).UpdateStatus(ctx, runCompletionEvent)).NotTo(Succeed())
			})
		})
	})

	Context("RunConfigurations", func() {
		var LastSucceededRunHasBeenUpdated = func() func(pipelinesv1.RunConfiguration, common.RunCompletionEvent, pipelinesv1.RunConfiguration) {
			return func(oldRun pipelinesv1.RunConfiguration, event common.RunCompletionEvent, newRun pipelinesv1.RunConfiguration) {
				Expect(newRun.Status.LatestRuns.Succeeded.ProviderId).To(Equal(event.RunId))
				Expect(newRun.Status.LatestRuns.Succeeded.Artifacts).To(BeComparableTo(event.Artifacts, cmpopts.EquateEmpty()))
			}
		}

		var HasNotChanged = func() func(pipelinesv1.RunConfiguration, common.RunCompletionEvent, pipelinesv1.RunConfiguration) {
			return func(oldRun pipelinesv1.RunConfiguration, event common.RunCompletionEvent, newRun pipelinesv1.RunConfiguration) {
				Expect(newRun.Status).To(Equal(oldRun.Status))
			}
		}

		DescribeTable("updates RunConfiguration on known states only",
			func(status common.RunCompletionStatus, expectation func(pipelinesv1.RunConfiguration, common.RunCompletionEvent, pipelinesv1.RunConfiguration)) {
				ctx := context.Background()
				runConfiguration := pipelinesv1.RandomRunConfiguration(apis.RandomLowercaseString())
				Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

				runCompletionEvent := common.RunCompletionEvent{
					Status: status,
					RunConfigurationName: &common.NamespacedName{
						Name:      runConfiguration.Name,
						Namespace: runConfiguration.Namespace,
					},
					RunId: common.RandomString(),
					Artifacts: apis.RandomList(func() common.Artifact {
						return common.Artifact{
							Name:     common.RandomString(),
							Location: common.RandomString(),
						}
					}),
				}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())

				fetchedRunConfiguration := pipelinesv1.RunConfiguration{}
				Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), &fetchedRunConfiguration)).To(Succeed())
				expectation(*runConfiguration, runCompletionEvent, fetchedRunConfiguration)
			},
			Entry("succeeded should succeed", common.RunCompletionStatuses.Succeeded, LastSucceededRunHasBeenUpdated()),
			Entry("failed should fail", common.RunCompletionStatuses.Failed, HasNotChanged()),
			Entry("unknown should not override", common.RunCompletionStatus(""), HasNotChanged()))

		When("RunConfiguration is not found", func() {
			It("do nothing", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunConfigurationName: &common.NamespacedName{
					Name:      common.RandomString(),
					Namespace: common.RandomString(),
				}}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())
			})
		})

		When("the runConfigurationName has no namespace", func() {
			It("do nothing", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunConfigurationName: &common.NamespacedName{
					Name: common.RandomString(),
				}}

				Expect(statusUpdater.UpdateStatus(ctx, runCompletionEvent)).To(Succeed())
			})
		})

		When("the k8s API is unreachable", func() {
			It("errors", func() {
				ctx := context.Background()

				runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunConfigurationName: &common.NamespacedName{
					Name:      common.RandomString(),
					Namespace: common.RandomString(),
				}}

				Expect((&StatusUpdater{
					NewFailingClient(),
				}).UpdateStatus(ctx, runCompletionEvent)).NotTo(Succeed())
			})
		})
	})
})
