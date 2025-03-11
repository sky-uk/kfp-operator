//go:build integration

package workflowfactory

import (
	"context"
	"testing"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	"github.com/sky-uk/kfp-operator/external"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TestTimeout   = 120
	TestNamespace = "argo"
	TestProvider  = "stub"
)

var (
	restCfg = rest.Config{
		Host:    "http://localhost:8080",
		APIPath: "/api",
	}
	TestProviderConfig = pipelinesv1.RandomProvider()
)

func TestPipelineControllersIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers Integration Suite")
}

var _ = BeforeSuite(func() {
	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())
	var err error
	K8sClient, err = client.New(&restCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Ctx = context.Background()
})

var provider = pipelinesv1.Provider{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "kfp-operator-integration-tests-providers",
		Namespace: TestNamespace,
	},
	Spec: pipelinesv1.ProviderSpec{
		ServiceImage:   "kfp-operator-stub-provider-service",
		Image:          "kfp-operator-unused-old-stub-provider-cli",
		ExecutionMode:  "none",
		ServiceAccount: "default",
	},
	Status: pipelinesv1.Status{},
}

func AssertWorkflow[R pipelinesv1.Resource](
	resource R,
	expectedOutput base.Output,
	constructWorkflow func(pipelinesv1.Provider, corev1.Service, R) (*argo.Workflow, error),
) {

	testCtx := WorkflowTestHelper[R]{
		Resource: resource,
	}

	providerSvc := corev1.Service{}
	err := K8sClient.Get(
		Ctx,
		types.NamespacedName{
			Namespace: TestNamespace,
			Name:      "provider-test",
		},
		&providerSvc,
	)
	Expect(err).ToNot(HaveOccurred())

	workflow, err := constructWorkflow(provider, providerSvc, testCtx.Resource)

	Expect(err).NotTo(HaveOccurred())
	Expect(K8sClient.Create(Ctx, workflow)).To(Succeed())

	Eventually(
		testCtx.WorkflowByNameToMatch(
			types.NamespacedName{
				Name:      workflow.Name,
				Namespace: workflow.Namespace,
			},
			func(g Gomega, workflow *argo.Workflow) {
				g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
				output, err := workflowutil.GetWorkflowOutput(
					workflow,
					workflowconstants.ProviderOutputParameterName,
				)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal(expectedOutput))
			},
		), TestTimeout,
	).Should(Succeed())
}

func withIntegrationTestFields[T pipelinesv1.Resource](resource T) T {
	resource.SetNamespace(TestNamespace)
	resourceStatus := resource.GetStatus()
	resourceStatus.Provider.Name = TestProvider
	resource.SetStatus(resourceStatus)

	return resource
}
