//go:build decoupled
// +build decoupled

package run_completer

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var (
	k8sClient    client.Client
	runCompleter RunCompleter
	testEnv      *envtest.Environment
)

func TestModelUpdateEventSourceDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventing Decoupled Suite")
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

	runCompleter = RunCompleter{
		K8sClient: k8sClient,
	}
})

var _ = AfterSuite(func() {
	testEnv.Stop()
})

func HasChangedTo(expectedState pipelinesv1.CompletionState) func(pipelinesv1.Run, pipelinesv1.Run) {
	return func(oldRun pipelinesv1.Run, newRun pipelinesv1.Run) {
		Expect(oldRun.Status.CompletionState).NotTo(Equal(expectedState))
		Expect(newRun.Status.CompletionState).To(Equal(expectedState))
	}
}

func HasNotChanged() func(pipelinesv1.Run, pipelinesv1.Run) {
	return func(oldRun pipelinesv1.Run, newRun pipelinesv1.Run) {
		Expect(newRun.Status).To(Equal(oldRun.Status))
	}
}

var _ = Context("Run Completer", Serial, func() {
	DescribeTable("updates Run on known states only",
		func(status common.RunCompletionStatus, expectation func(pipelinesv1.Run, pipelinesv1.Run)) {
			ctx := context.Background()
			run := pipelinesv1.RandomRun()
			Expect(k8sClient.Create(ctx, run)).To(Succeed())

			runCompletionEvent := common.RunCompletionEvent{Status: status, RunName: &common.NamespacedName{
				Name:      run.Name,
				Namespace: run.Namespace,
			}}

			Expect(runCompleter.CompleteRun(ctx, runCompletionEvent)).To(Succeed())

			fetchedRun := pipelinesv1.Run{}
			Expect(k8sClient.Get(ctx, run.GetNamespacedName(), &fetchedRun)).To(Succeed())
			expectation(*run, fetchedRun)
		},
		Entry("succeeded should succeed", common.RunCompletionStatuses.Succeeded, HasChangedTo(pipelinesv1.CompletionStates.Succeeded)),
		Entry("failed should fail", common.RunCompletionStatuses.Failed, HasChangedTo(pipelinesv1.CompletionStates.Failed)),
		Entry("unknown should not override", common.RunCompletionStatus(""), HasNotChanged()))

	When("the run is not found", func() {
		It("do nothing", func() {
			ctx := context.Background()

			runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
				Name:      common.RandomString(),
				Namespace: common.RandomString(),
			}}

			Expect(runCompleter.CompleteRun(ctx, runCompletionEvent)).To(Succeed())
		})
	})

	When("the run name has no namespace", func() {
		It("do nothing", func() {
			ctx := context.Background()

			runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
				Name:      common.RandomString(),
			}}

			Expect(runCompleter.CompleteRun(ctx, runCompletionEvent)).To(Succeed())
		})
	})

	When("the k8s API is unreachable", func() {
		It("errors", func() {
			ctx := context.Background()

			runCompletionEvent := common.RunCompletionEvent{Status: common.RunCompletionStatuses.Succeeded, RunName: &common.NamespacedName{
				Name:      common.RandomString(),
				Namespace: common.RandomString(),
			}}

			Expect((&RunCompleter{
				NewFailingClient(),
			}).CompleteRun(ctx, runCompletionEvent)).NotTo(Succeed())
		})
	})
})
