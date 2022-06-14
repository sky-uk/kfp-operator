//go:build decoupled
// +build decoupled

package pipelines

import (
	"context"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/external"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"

	. "github.com/onsi/ginkgo/v2"
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
	//logf.SetLogger(logf.NullLogSink{})

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

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
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

	Expect((NewTestPipelineReconciler(ec, &workflowRepository)).SetupWithManager(k8sManager)).To(Succeed())
	Expect((NewTestRunConfigurationReconciler(ec, &workflowRepository)).SetupWithManager(k8sManager)).To(Succeed())
	Expect((NewTestExperimentReconciler(ec, &workflowRepository)).SetupWithManager(k8sManager)).To(Succeed())
	Expect(workflowRepository.SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
})

func NewTestPipelineReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *PipelineReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = PipelineWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage:   "kfp-sdk",
					CompilerImage: "compiler",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never",
					},
				},
				KfpEndpoint: "http://www.example.com",
			},
		},
	}

	var stateHandler = PipelineStateHandler{
		WorkflowRepository: workflowRepository,
		WorkflowFactory:    workflowFactory,
	}

	return &PipelineReconciler{
		EC:           ec,
		StateHandler: stateHandler,
	}
}

func NewTestRunConfigurationReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *RunConfigurationReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = RunConfigurationWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				DefaultExperiment: "Default",
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage:   "kfp-sdk",
					CompilerImage: "compiler",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never",
					},
				},
				KfpEndpoint: "http://www.example.com",
			},
		},
	}

	var stateHandler = RunConfigurationStateHandler{
		WorkflowRepository: workflowRepository,
		WorkflowFactory:    workflowFactory,
	}

	return &RunConfigurationReconciler{
		EC:           ec,
		StateHandler: stateHandler,
	}
}

func NewTestExperimentReconciler(ec K8sExecutionContext, workflowRepository WorkflowRepository) *ExperimentReconciler {
	// TODO: mock workflowFactory
	var workflowFactory = ExperimentWorkflowFactory{
		WorkflowFactory: WorkflowFactory{
			Config: configv1.Configuration{
				Argo: configv1.ArgoConfiguration{
					KfpSdkImage:   "kfp-sdk",
					CompilerImage: "compiler",
					ContainerDefaults: apiv1.Container{
						ImagePullPolicy: "Never",
					},
				},
				KfpEndpoint: "http://www.example.com",
			},
		},
	}

	var stateHandler = ExperimentStateHandler{
		WorkflowRepository: workflowRepository,
		WorkflowFactory:    workflowFactory,
	}

	return &ExperimentReconciler{
		EC:           ec,
		StateHandler: stateHandler,
	}
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
