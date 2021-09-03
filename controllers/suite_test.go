package controllers

import (
	"context"
	"path/filepath"
	"testing"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thanhpk/randstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	//+kubebuilder:scaffold:imports

	"github.com/sky-uk/kfp-operator/external"
)

const (
	PipelineNamespace = "default"
	PipelineId        = "12345"
	AnotherPipelineId = "67890"
)

type TestContext struct {
	Pipeline          *pipelinesv1.Pipeline
	PipelineLookupKey types.NamespacedName
	Version           string
}

var specV1 = pipelinesv1.PipelineSpec{
	Image:         "image:v1",
	TfxComponents: "pipeline.create_components",
	Env: map[string]string{
		"a": "aVal",
		"b": "bVal",
	},
}

var specV2 = pipelinesv1.PipelineSpec{
	Image:         "image:v1",
	TfxComponents: "pipeline.create_components",
	Env: map[string]string{
		"a": "aVal",
		"b": "bVal",
		"c": "cVal",
	},
}

func randomPipeline() *pipelinesv1.Pipeline {
	return &pipelinesv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      randstr.String(16, "0123456789abcdefghijklmnopqrstuvwxyz"),
			Namespace: PipelineNamespace,
		},
		Spec: pipelinesv1.PipelineSpec{
			Image:         "image:v1",
			TfxComponents: "pipeline.create_components",
			Env: map[string]string{
				"a": "aVal",
				"b": "bVal",
			},
		},
	}
}

var v0 = pipelinesv1.ComputeVersion(pipelinesv1.PipelineSpec{})
var v1 = pipelinesv1.ComputeVersion(specV1)
var v2 = pipelinesv1.ComputeVersion(specV2)

func NewTestContext() TestContext {
	pipeline := randomPipeline()

	return TestContext{
		Pipeline:          pipeline,
		PipelineLookupKey: types.NamespacedName{Name: pipeline.ObjectMeta.Name, Namespace: PipelineNamespace},
		Version:           pipelinesv1.ComputeVersion(pipeline.Spec),
	}
}

func (testCtx TestContext) pipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		g.Expect(k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline)).To(Succeed())

		matcher(g, pipeline)
	}
}

func (testCtx TestContext) pipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline)

	return err
}

func (testCtx TestContext) workflowInputToMatch(operation string, matcher func(Gomega, map[string]string)) func(Gomega) {

	var mapParams = func(params []argo.Parameter) map[string]string {
		m := make(map[string]string, len(params))
		for i := range params {
			m[params[i].Name] = string(*params[i].Value)
		}

		return m
	}

	return func(g Gomega) {
		workflows := &argo.WorkflowList{}
		g.Expect(k8sClient.List(ctx, workflows, client.InNamespace(testCtx.Pipeline.ObjectMeta.Namespace), client.MatchingLabels{OperationLabelKey: operation, PipelineLabelKey: testCtx.Pipeline.ObjectMeta.Name})).To(Succeed())
		g.Expect(len(workflows.Items)).To(Equal(1))

		worklfowInputParameters := mapParams(workflows.Items[0].Spec.Arguments.Parameters)
		matcher(g, worklfowInputParameters)
	}
}

func (testCtx TestContext) updateWorkflow(operation string, updateFunc func(*argo.Workflow)) error {
	workflows := &argo.WorkflowList{}
	if err := k8sClient.List(ctx, workflows, client.InNamespace(testCtx.Pipeline.ObjectMeta.Namespace), client.MatchingLabels{OperationLabelKey: operation, PipelineLabelKey: testCtx.Pipeline.ObjectMeta.Name}); err != nil {
		return err
	}

	Expect(len(workflows.Items)).To(Equal(1))

	updateFunc(&workflows.Items[0])
	return k8sClient.Update(ctx, &workflows.Items[0])
}

func (testCtx TestContext) updatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return k8sClient.Update(ctx, pipeline)
}

func (testCtx TestContext) updatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return k8sClient.Status().Update(ctx, pipeline)
}

func (testCtx TestContext) pipelineCreated() {
	testCtx.pipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
		Id:                   PipelineId,
		SynchronizationState: pipelinesv1.Succeeded,
		Version:              testCtx.Version,
	})
}

func (testCtx TestContext) deletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := k8sClient.Get(ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	return k8sClient.Delete(ctx, pipeline)
}

func (testCtx TestContext) pipelineCreatedWithStatus(status pipelinesv1.PipelineStatus) {
	Expect(k8sClient.Create(ctx, testCtx.Pipeline)).To(Succeed())

	Eventually(testCtx.pipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
		g.Expect(testCtx.updatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}

var cfg *rest.Config
var k8sClient client.Client
var ctx context.Context
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "config", "crd", "external"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = pipelinesv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	ctx = context.Background()

	Expect((&PipelineReconciler{
		Client: k8sClient,
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
