package pipelines

import (
	"context"
	"errors"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	//+kubebuilder:scaffold:imports
)

const (
	PipelineNamespace = "default"
	PipelineId        = "12345"
	AnotherPipelineId = "67890"
)

type TestContext struct {
	K8sClient         client.Client
	ctx               context.Context
	Pipeline          *pipelinesv1.Pipeline
	PipelineLookupKey types.NamespacedName
	Version           string
}

var SpecV1 = pipelinesv1.PipelineSpec{
	Image:         "test-pipeline",
	TfxComponents: "pipeline.create_components",
	Env: map[string]string{
		"a": "aVal",
		"b": "bVal",
	},
}

var SpecV2 = pipelinesv1.PipelineSpec{
	Image:         "test-pipeline",
	TfxComponents: "pipeline.create_components",
	Env: map[string]string{
		"a": "aVal",
		"b": "bVal",
		"c": "cVal",
	},
}

var V0 = pipelinesv1.PipelineSpec{}.ComputeVersion()
var V1 = SpecV1.ComputeVersion()
var V2 = SpecV2.ComputeVersion()

func NewTestContextWithPipeline(pipeline *pipelinesv1.Pipeline, k8sClient client.Client, ctx context.Context) TestContext {
	return TestContext{
		K8sClient:         k8sClient,
		ctx:               ctx,
		Pipeline:          pipeline,
		PipelineLookupKey: types.NamespacedName{Name: pipeline.Name, Namespace: PipelineNamespace},
		Version:           pipeline.Spec.ComputeVersion(),
	}
}

func NewTestContext(k8sClient client.Client, ctx context.Context) TestContext {
	return NewTestContextWithPipeline(RandomPipeline(), k8sClient, ctx)
}

func (testCtx TestContext) PipelineToMatch(matcher func(Gomega, *pipelinesv1.Pipeline)) func(Gomega) {
	return func(g Gomega) {
		pipeline := &pipelinesv1.Pipeline{}
		Expect(testCtx.K8sClient.Get(testCtx.ctx, testCtx.PipelineLookupKey, pipeline)).To(Succeed())
		matcher(g, pipeline)
	}
}

func (testCtx TestContext) PipelineExists() error {
	pipeline := &pipelinesv1.Pipeline{}
	err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.PipelineLookupKey, pipeline)

	return err
}

func (testCtx TestContext) WorkflowInputToMatch(operation string, matcher func(Gomega, map[string]string)) func(Gomega) {

	var mapParams = func(params []argo.Parameter) map[string]string {
		m := make(map[string]string, len(params))
		for i := range params {
			m[params[i].Name] = string(*params[i].Value)
		}

		return m
	}

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		worklfowInputParameters := mapParams(workflow.Spec.Arguments.Parameters)
		matcher(g, worklfowInputParameters)
	}
}

func (testCtx TestContext) WorkflowToMatch(operation string, matcher func(Gomega, *argo.Workflow)) func(Gomega) {

	return func(g Gomega) {
		workflow, err := testCtx.fetchWorkflow(operation)

		Expect(err).NotTo(HaveOccurred())

		matcher(g, workflow)
	}
}

func (testCtx TestContext) UpdateWorkflow(operation string, updateFunc func(*argo.Workflow)) error {
	workflow, err := testCtx.fetchWorkflow(operation)
	if err != nil {
		return err
	}

	updateFunc(workflow)
	return testCtx.K8sClient.Update(testCtx.ctx, workflow)
}

func (testCtx TestContext) fetchWorkflow(operation string) (*argo.Workflow, error) {
	workflowList := &argo.WorkflowList{}

	if err := testCtx.K8sClient.List(testCtx.ctx, workflowList, client.MatchingLabels{OperationLabelKey: operation, PipelineNameLabelKey: testCtx.Pipeline.Name}); err != nil {
		return nil, err
	}

	if len(workflowList.Items) != 1 {
		return nil, errors.New("not exactly one workflow")
	}

	return &workflowList.Items[0], nil
}

func (testCtx TestContext) UpdatePipeline(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Update(testCtx.ctx, pipeline)
}

func (testCtx TestContext) UpdatePipelineStatus(updateFunc func(*pipelinesv1.Pipeline)) error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	updateFunc(pipeline)

	return testCtx.K8sClient.Status().Update(testCtx.ctx, pipeline)
}

func (testCtx TestContext) PipelineCreated() {
	testCtx.PipelineCreatedWithStatus(pipelinesv1.PipelineStatus{
		Id:                   PipelineId,
		SynchronizationState: pipelinesv1.Succeeded,
		Version:              testCtx.Version,
	})
}

func (testCtx TestContext) DeletePipeline() error {
	pipeline := &pipelinesv1.Pipeline{}

	if err := testCtx.K8sClient.Get(testCtx.ctx, testCtx.PipelineLookupKey, pipeline); err != nil {
		return err
	}

	return testCtx.K8sClient.Delete(testCtx.ctx, pipeline)
}

func (testCtx TestContext) PipelineCreatedWithStatus(status pipelinesv1.PipelineStatus) {
	Expect(testCtx.K8sClient.Create(testCtx.ctx, testCtx.Pipeline)).To(Succeed())

	Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
		g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
		g.Expect(testCtx.UpdatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
			pipeline.Status = status
		})).To(Succeed())
	})).Should(Succeed())
}
