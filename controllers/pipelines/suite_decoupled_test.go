//go:build decoupled

package pipelines

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/external"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func TestPipelineControllersDecoupledSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers Decoupled Suite")
}

var (
	testEnv    *envtest.Environment
	k8sManager manager.Manager
)

const (
	TestFramework = "dummyframework"
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

	err = pipelineshub.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())

	K8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(K8sClient).NotTo(BeNil())

	webhookInstallOptions := &testEnv.WebhookInstallOptions
	webhook := webhook.NewServer(webhook.Options{
		Host:    webhookInstallOptions.LocalServingHost,
		Port:    webhookInstallOptions.LocalServingPort,
		CertDir: webhookInstallOptions.LocalServingCertDir,
	})
	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:         scheme.Scheme,
		WebhookServer:  webhook,
		LeaderElection: false,
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
		Scheme:             k8sManager.GetScheme(),
		WorkflowRepository: workflowRepository,
	}

	Ctx = context.Background()

	TestConfig = config.KfpControllerConfigSpec{
		DefaultProviderValues: config.DefaultProviderValues{
			ServicePort: 8080,
		},
		DefaultExperiment: "Default",
		WorkflowNamespace: "default",
		DefaultProvider:   apis.RandomLowercaseString(),
		RunCompletionTTL:  &metav1.Duration{Duration: time.Minute},
	}

	Expect(NewPipelineReconciler(ec, &workflowRepository, TestConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunReconciler(ec, &workflowRepository, TestConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunConfigurationReconciler(ec, TestConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunScheduleReconciler(ec, &workflowRepository, TestConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewExperimentReconciler(ec, &workflowRepository, TestConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect((&pipelineshub.RunConfiguration{}).SetupWebhookWithManager(k8sManager)).To(Succeed())
	Expect((&pipelineshub.Run{}).SetupWebhookWithManager(k8sManager)).To(Succeed())

	Provider = pipelineshub.RandomProvider()
	Provider.Name = apis.RandomLowercaseString()
	Provider.Namespace = TestConfig.WorkflowNamespace
	Provider.Spec.Frameworks = []pipelineshub.Framework{
		{
			Name:  TestFramework,
			Image: "dummy-image",
		},
	}
	Expect(K8sClient.Create(Ctx, Provider)).To(Succeed())

	providerSvc := corev1.Service{}
	providerSvc.SetName("provider-svc")
	providerSvc.SetNamespace(Provider.Namespace)
	providerSvc.Spec.Ports = []corev1.ServicePort{{Port: 8080}}
	providerSvc.SetOwnerReferences(
		[]metav1.OwnerReference{
			*metav1.NewControllerRef(
				Provider,
				pipelineshub.GroupVersion.WithKind(Provider.GetKind()),
			),
		},
	)
	Expect(K8sClient.Create(Ctx, &providerSvc)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
})

var _ = BeforeEach(func() {
	allRuns := &pipelineshub.RunList{}
	Expect(K8sClient.List(Ctx, allRuns)).To(Succeed())
	for _, r := range allRuns.Items {
		Expect(client.IgnoreNotFound(K8sClient.Delete(Ctx, &r))).To(Succeed())
	}

	allRunSchedules := &pipelineshub.RunScheduleList{}
	Expect(K8sClient.List(Ctx, allRunSchedules)).To(Succeed())
	for _, r := range allRunSchedules.Items {
		Expect(client.IgnoreNotFound(K8sClient.Delete(Ctx, &r))).To(Succeed())
	}

	allRcs := &pipelineshub.RunConfigurationList{}
	Expect(K8sClient.List(Ctx, allRcs)).To(Succeed())
	for _, r := range allRcs.Items {
		Expect(client.IgnoreNotFound(K8sClient.Delete(Ctx, &r))).To(Succeed())
	}

	allPipelines := &pipelineshub.PipelineList{}
	Expect(K8sClient.List(Ctx, allPipelines)).To(Succeed())
	for _, r := range allPipelines.Items {
		Expect(client.IgnoreNotFound(K8sClient.Delete(Ctx, &r))).To(Succeed())
	}
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
