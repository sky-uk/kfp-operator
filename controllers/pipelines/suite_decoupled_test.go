//go:build decoupled
// +build decoupled

package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/external"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"
)

func TestPipelineControllersDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers Decoupled Suite")
}

var (
	testEnv    *envtest.Environment
	k8sManager manager.Manager
)

var _ = BeforeSuite(func() {

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "config", "crd", "external"),
		},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = pipelinesv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	webhookInstallOptions := &testEnv.WebhookInstallOptions
	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	optInClient := controllers.NewOptInClient(k8sManager)

	var workflowRepository = WorkflowRepositoryImpl{
		Client: optInClient,
		Scheme: k8sManager.GetScheme(),
	}

	ec := K8sExecutionContext{
		Client:             optInClient,
		Recorder:           k8sManager.GetEventRecorderFor("decoupled-test-controller"),
		WorkflowRepository: workflowRepository,
	}

	ctx = context.Background()

	testConfig = config.Configuration{
		DefaultExperiment: "Default",
		WorkflowNamespace: "default",
		DefaultProvider:   apis.RandomString(),
		RunCompletionTTL:  &metav1.Duration{Duration: time.Minute},
	}

	Expect(NewTestPipelineReconciler(ec, &workflowRepository).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewTestRunReconciler(ec, &workflowRepository).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewTestRunConfigurationReconciler(ec, &workflowRepository).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewTestExperimentReconciler(ec, &workflowRepository).SetupWithManager(k8sManager)).To(Succeed())
	Expect(workflowRepository.SetupWithManager(k8sManager)).To(Succeed())
	Expect((&pipelinesv1.Run{}).SetupWebhookWithManager(k8sManager)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
})

func NewTestPipelineReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *PipelineReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = PipelineWorkflowFactory(testConfig)

	return &PipelineReconciler{
		BaseReconciler: BaseReconciler[*pipelinesv1.Pipeline]{
			Config: testConfig,
			EC:     ec,
			StateHandler: StateHandler[*pipelinesv1.Pipeline]{
				WorkflowRepository: workflowRepository,
				WorkflowFactory:    &workflowFactory,
			},
		},
	}
}

func NewTestRunReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *RunReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = RunWorkflowFactory(testConfig)

	return &RunReconciler{
		BaseReconciler: BaseReconciler[*pipelinesv1.Run]{
			Config: testConfig,
			EC:     ec,
			StateHandler: StateHandler[*pipelinesv1.Run]{
				WorkflowRepository: workflowRepository,
				WorkflowFactory:    &workflowFactory,
			},
		},
	}
}

func NewTestRunConfigurationReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *RunConfigurationReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = RunConfigurationWorkflowFactory(testConfig)

	return &RunConfigurationReconciler{
		BaseReconciler: BaseReconciler[*pipelinesv1.RunConfiguration]{
			Config: testConfig,
			EC:     ec,
			StateHandler: StateHandler[*pipelinesv1.RunConfiguration]{
				WorkflowRepository: workflowRepository,
				WorkflowFactory:    &workflowFactory,
			},
		},
	}
}

func NewTestExperimentReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *ExperimentReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = ExperimentWorkflowFactory(testConfig)

	return &ExperimentReconciler{
		BaseReconciler: BaseReconciler[*pipelinesv1.Experiment]{
			Config: testConfig,
			EC:     ec,
			StateHandler: StateHandler[*pipelinesv1.Experiment]{
				WorkflowRepository: workflowRepository,
				WorkflowFactory:    &workflowFactory,
			},
		},
	}
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
