//go:build decoupled
// +build decoupled

package pipelines

import (
	"context"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/external"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	var workflowRepository = WorkflowRepositoryImpl{
		Client: k8sClient,
	}

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	ctx = context.Background()

	Expect((NewTestPipelineReconciler(k8sManager, &workflowRepository)).SetupWithManager(k8sManager)).To(Succeed())
	Expect((NewTestRunConfigurationReconciler(k8sManager, &workflowRepository)).SetupWithManager(k8sManager)).To(Succeed())
	Expect(workflowRepository.SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
}, 60)

func NewTestPipelineReconciler(k8sManager manager.Manager, workflowRepository WorkflowRepository) *PipelineReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = PipelineWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpSdkImage:     "kfp-sdk",
				CompilerImage:   "compiler",
				ImagePullPolicy: "Never",
				KfpEndpoint:     "http://www.example.com",
			},
		},
	}

	var stateHandler = PipelineStateHandler{
		WorkflowRepository: workflowRepository,
		WorkflowFactory:    workflowFactory,
	}

	return &PipelineReconciler{
		Client:       k8sClient,
		Scheme:       k8sManager.GetScheme(),
		StateHandler: stateHandler,
	}
}

func NewTestRunConfigurationReconciler(k8sManager manager.Manager, workflowRepository WorkflowRepository) *RunConfigurationReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = RunConfigurationWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				KfpSdkImage:     "kfp-sdk",
				CompilerImage:   "compiler",
				ImagePullPolicy: "Never",
				KfpEndpoint:     "http://www.example.com",
			},
		},
	}

	var stateHandler = RunConfigurationStateHandler{
		WorkflowRepository: workflowRepository,
		WorkflowFactory:    workflowFactory,
	}

	return &RunConfigurationReconciler{
		Client:       k8sClient,
		Scheme:       k8sManager.GetScheme(),
		StateHandler: stateHandler,
	}
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
