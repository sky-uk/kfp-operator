//go:build integration

package pipelines

import (
	"context"
	"encoding/json"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/stub"
	"github.com/sky-uk/kfp-operator/external"
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
	k8sClient, err = client.New(&restCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	ctx = context.Background()
})

var _ = BeforeEach(func() {
	k8sClient.Delete(ctx, &pipelinesv1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfp-operator-integration-tests-providers",
			Namespace: TestNamespace,
		}})
})

func StubProvider[R pipelinesv1.Resource](stubbedOutput base.Output, resource R) (pipelinesv1.Provider, base.Output) {
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

	provider := pipelinesv1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfp-operator-integration-tests-providers",
			Namespace: TestNamespace,
		},
		Spec: pipelinesv1.ProviderSpec{
			Image:          "kfp-operator-stub-provider",
			ExecutionMode:  "none",
			ServiceAccount: "default",
			Parameters: map[string]*apiextensionsv1.JSON{
				"expectedInput":  {Raw: expectedInputJson},
				"expectedOutput": {Raw: expectedOutputJson},
			},
		},
		Status: pipelinesv1.ProviderStatus{},
	}
	return provider, stubbedOutput
}

func StubWithIdAndError[R pipelinesv1.Resource](resource R) (pipelinesv1.Provider, base.Output) {
	return StubProvider(base.Output{
		Id:            apis.RandomString(),
		ProviderError: apis.RandomString(),
	}, resource)
}

func StubWithEmpty[R pipelinesv1.Resource](resource R) (pipelinesv1.Provider, base.Output) {
	return StubProvider(base.Output{}, resource)
}

func StubWithExistingIdAndError[R pipelinesv1.Resource](resource R) (pipelinesv1.Provider, base.Output) {
	return StubProvider(base.Output{
		Id:            resource.GetStatus().Provider.Id,
		ProviderError: apis.RandomString(),
	}, resource)
}

func AssertWorkflow[R pipelinesv1.Resource](
	newResource func() R,
	setUp func(resource R) (pipelinesv1.Provider, base.Output),
	constructWorkflow func(pipelinesv1.Provider, R) (*argo.Workflow, error)) {

	testCtx := WorkflowTestHelper[R]{
		Resource: newResource(),
	}

	provider, expectedOutput := setUp(testCtx.Resource)
	workflow, err := constructWorkflow(provider, testCtx.Resource)

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient.Create(ctx, workflow)).To(Succeed())

	Eventually(testCtx.WorkflowByNameToMatch(types.NamespacedName{Name: workflow.Name, Namespace: workflow.Namespace},
		func(g Gomega, workflow *argo.Workflow) {
			g.Expect(workflow.Status.Phase).To(Equal(argo.WorkflowSucceeded))
			output, err := getWorkflowOutput(workflow, WorkflowConstants.ProviderOutputParameterName)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(output.ProviderError).To(Equal(expectedOutput.ProviderError))
			g.Expect(output.Id).To(Equal(expectedOutput.Id))
		}), TestTimeout).Should(Succeed())
}

func withIntegrationTestFields[T pipelinesv1.Resource](resource T) T {
	resource.SetNamespace(TestNamespace)
	resourceStatus := resource.GetStatus()
	resourceStatus.Provider.Name = TestProvider
	resource.SetStatus(resourceStatus)

	return resource
}
