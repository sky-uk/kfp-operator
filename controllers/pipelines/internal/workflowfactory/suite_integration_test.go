//go:build integration

package workflowfactory

import (
	"context"
	"encoding/json"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/stub"
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
	TestProviderConfig = pipelineshub.RandomProvider()
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

func StubProvider[R pipelineshub.Resource](
	stubbedOutput base.Output,
	resource R,
) (pipelineshub.Provider, base.Output) {
	expectedInput := stub.ExpectedInput{
		Id: resource.GetStatus().Provider.Id,
		ResourceDefinition: stub.ResourceDefinition{
			Name: common.NamespacedName{
				Name:      resource.GetName(),
				Namespace: resource.GetNamespace(),
			},
			Version: resource.ComputeVersion(),
		},
	}

	expectedInputJson, err := json.Marshal(expectedInput)
	Expect(err).NotTo(HaveOccurred())

	expectedOutputJson, err := json.Marshal(stubbedOutput)
	Expect(err).NotTo(HaveOccurred())

	provider := pipelineshub.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfp-operator-integration-tests-providers",
			Namespace: TestNamespace,
		},
		Spec: pipelineshub.ProviderSpec{
			ServiceImage:   "kfp-operator-stub-provider",
			Image:          "kfp-operator-stub-provider",
			ExecutionMode:  "none",
			ServiceAccount: "default",
			Parameters: map[string]*apiextensionsv1.JSON{
				"expectedInput":  {Raw: expectedInputJson},
				"expectedOutput": {Raw: expectedOutputJson},
			},
		},
		Status: pipelineshub.Status{},
	}
	return provider, stubbedOutput
}

var provider = pipelineshub.Provider{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "kfp-operator-integration-tests-providers",
		Namespace: TestNamespace,
	},
	Spec: pipelineshub.ProviderSpec{
		ServiceImage:   "kfp-operator-stub-provider",
		Image:          "kfp-operator-stub-provider",
		ExecutionMode:  "none",
		ServiceAccount: "default",
	},
	Status: pipelineshub.Status{},
}

func StubWithIdAndError[R pipelineshub.Resource](resource R) (pipelinesv1.Provider, base.Output) {
	return StubProvider(base.Output{
		Id:            apis.RandomString(),
		ProviderError: apis.RandomString(),
	}, resource)
}

func StubWithEmpty[R pipelineshub.Resource](resource R) (pipelineshub.Provider, base.Output) {
	return StubProvider(base.Output{}, resource)
}

func StubWithExistingIdAndError[R pipelineshub.Resource](resource R) (pipelineshub.Provider, base.Output) {
	return StubProvider(base.Output{
		Id:            resource.GetStatus().Provider.Id,
		ProviderError: apis.RandomString(),
	}, resource)
}

func AssertWorkflow[R pipelineshub.Resource](
	resource R,
	expectedOutput base.Output,
	constructWorkflow func(pipelineshub.Provider, corev1.Service, R) (*argo.Workflow, error),
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

func withIntegrationTestFields[T pipelineshub.Resource](resource T) T {
	resource.SetNamespace(TestNamespace)
	resourceStatus := resource.GetStatus()
	resourceStatus.Provider.Name = TestProvider
	resource.SetStatus(resourceStatus)

	return resource
}
