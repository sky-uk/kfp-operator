//go:build decoupled

package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
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
		Scheme:             k8sManager.GetScheme(),
		WorkflowRepository: workflowRepository,
	}

	ctx = context.Background()

	testConfig = config.Configuration{
		DefaultExperiment: "Default",
		WorkflowNamespace: "default",
		DefaultProvider:   apis.RandomString(),
		RunCompletionTTL:  &metav1.Duration{Duration: time.Minute},
	}

	Expect(NewPipelineReconciler(ec, &workflowRepository, testConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunReconciler(ec, &workflowRepository, testConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunConfigurationReconciler(ec, k8sManager.GetScheme(), testConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewRunScheduleReconciler(ec, &workflowRepository, testConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect(NewExperimentReconciler(ec, &workflowRepository, testConfig).SetupWithManager(k8sManager)).To(Succeed())
	Expect((&pipelinesv1.RunConfiguration{}).SetupWebhookWithManager(k8sManager)).To(Succeed())
	Expect((&pipelinesv1.Run{}).SetupWebhookWithManager(k8sManager)).To(Succeed())

	go func() {
		Expect(k8sManager.Start(ctrl.SetupSignalHandler())).To(Succeed())
	}()
})

var _ = BeforeEach(func() {
	allRuns := &pipelinesv1.RunList{}
	Expect(k8sClient.List(ctx, allRuns)).To(Succeed())
	for _, r := range allRuns.Items {
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &r))).To(Succeed())
	}

	allRunSchedules := &pipelinesv1.RunScheduleList{}
	Expect(k8sClient.List(ctx, allRunSchedules)).To(Succeed())
	for _, r := range allRunSchedules.Items {
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &r))).To(Succeed())
	}

	allRcs := &pipelinesv1.RunConfigurationList{}
	Expect(k8sClient.List(ctx, allRcs)).To(Succeed())
	for _, r := range allRcs.Items {
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &r))).To(Succeed())
	}

	allPipelines := &pipelinesv1.PipelineList{}
	Expect(k8sClient.List(ctx, allPipelines)).To(Succeed())
	for _, r := range allPipelines.Items {
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &r))).To(Succeed())
	}
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
