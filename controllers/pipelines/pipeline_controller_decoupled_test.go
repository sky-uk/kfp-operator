// +build decoupled

package pipelines

import (
	"context"
	"path/filepath"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/external"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Pipeline controller k8s integration", func() {
	var k8sClient client.Client
	var ctx context.Context
	var testEnv *envtest.Environment

	BeforeSuite(func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("..", "..", "config", "crd", "bases"),
				filepath.Join("..", "..", "config", "crd", "external"),
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

		// TODO: mock workflowFactory
		var workflowFactory = WorkflowFactory{
			Config: configv1.Configuration{
				KfpToolsImage:   "kfp-tools",
				CompilerImage:   "compiler",
				ImagePullPolicy: "Never",
				KfpEndpoint:     "http://www.example.com",
			},
		}

		var workflowRepository = WorkflowRepositoryImpl{
			Client: k8sClient,
		}

		var stateHandler = StateHandler{
			WorkflowRepository: workflowRepository,
			WorkflowFactory:    workflowFactory,
		}

		var reconciler = PipelineReconciler{
			Client:       k8sClient,
			Scheme:       k8sManager.GetScheme(),
			StateHandler: stateHandler,
		}

		Expect((&reconciler).SetupWithManager(k8sManager)).To(Succeed())

		go func() {
			Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
		}()
	}, 60)

	AfterSuite(func() {
		By("tearing down the test environment")
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			pipeline := &pipelinesv1.Pipeline{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RandomLowercaseString(),
					Namespace: "default",
				},
				Spec: SpecV1,
			}

			testCtx := NewTestContextWithPipeline(pipeline, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.Pipeline)).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, WorkflowFactoryConstants.pipelineIdParameterName, PipelineId)
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.UpdatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = SpecV2
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.DeletePipeline()).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.PipelineExists).Should(Not(Succeed()))
		})
	})
})
