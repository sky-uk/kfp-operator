//go:build decoupled
// +build decoupled

package eventing

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

var (
	k8sClient client.Client
	runCompleter         RunCompleter
)

const (
	defaultNamespace = "default"
)

func TestModelUpdateEventSourceDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventing Decoupled Suite")
}

var _ = BeforeSuite(func() {
	testEnv := &envtest.Environment{
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

func HasChangedTo(expectedState pipelinesv1.CompletionState) func(pipelinesv1.Run, pipelinesv1.Run) {
	return func(_ pipelinesv1.Run, newRun  pipelinesv1.Run) {
		Expect(newRun.Status.CompletionState).To(Equal(expectedState))
	}
}

func HasNotChanged() func(pipelinesv1.Run, pipelinesv1.Run) {
	return func(oldRun pipelinesv1.Run, newRun  pipelinesv1.Run) {
		Expect(newRun.Status).To(Equal(oldRun.Status))
	}
}

var _ = Context("Run Completer", Serial, func() {
	DescribeTable("known states",
		func(status RunCompletionStatus, expectation func(pipelinesv1.Run, pipelinesv1.Run)) {
			ctx := context.Background()
			run := pipelinesv1.RandomRun()
			Expect(k8sClient.Create(ctx, run)).To(Succeed())

			runCompletionEvent := RunCompletionEvent{Status: status, Run: NamespacedName{
				Name: run.Name,
				Namespace: run.Namespace,
			}}

			Expect(runCompleter.CompleteRun(ctx, runCompletionEvent)).To(Succeed())

			fetchedRun := pipelinesv1.Run{}
			Expect(k8sClient.Get(ctx, run.GetNamespacedName(), &fetchedRun)).To(Succeed())
			expectation(*run, fetchedRun)
		},
		Entry("succeeded should succeed", RunCompletionStatuses.Succeeded, HasChangedTo(pipelinesv1.CompletionStates.Succeeded)),
		Entry("failed should fail", RunCompletionStatuses.Failed, HasChangedTo(pipelinesv1.CompletionStates.Failed)),
		Entry("unknown should not override", RunCompletionStatus(""), HasNotChanged()))
})
