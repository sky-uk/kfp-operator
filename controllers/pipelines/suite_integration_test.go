//go:build integration

package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/argo/providers/stub"
	"github.com/sky-uk/kfp-operator/external"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

const (
	TestTimeout = 120
)

var (
	restCfg = rest.Config{
		Host:    "http://localhost:8080",
		APIPath: "/api",
	}
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
	k8sClient.Delete(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfp-operator-integration-tests-providers",
			Namespace: "argo",
		}})
})

func StubProvider[R pipelinesv1.Resource](stubbedOutput base.Output, resource R) base.Output {
	providerConfig := stub.StubProviderConfig{
		StubbedOutput: stubbedOutput,
		ExpectedInput: stub.ExpectedInput{
			Id: resource.GetStatus().ProviderId.Id,
			ResourceDefinition: stub.ResourceDefinition{
				Name: common.NamespacedName{
					Name:      resource.GetName(),
					Namespace: resource.GetNamespace(),
				},
				Version: resource.ComputeVersion(),
			},
		},
	}

	configYaml, err := yaml.Marshal(providerConfig)
	Expect(err).NotTo(HaveOccurred())

	Expect(k8sClient.Create(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfp-operator-integration-tests-providers",
			Namespace: "argo",
		},
		Data: map[string]string{
			"stub": fmt.Sprintf("%s\nserviceAccount: default\nimage: kfp-operator-stub-provider\nexecutionMode: none", configYaml),
		},
	})).To(Succeed())

	return providerConfig.StubbedOutput
}

func StubWithIdAndError[R pipelinesv1.Resource](resource R) base.Output {
	return StubProvider(base.Output{
		Id:            apis.RandomString(),
		ProviderError: apis.RandomString(),
	}, resource)
}

func StubWithEmpty[R pipelinesv1.Resource](resource R) base.Output {
	return StubProvider(base.Output{}, resource)
}

func StubWithExistingIdAndError[R pipelinesv1.Resource](resource R) base.Output {
	return StubProvider(base.Output{
		Id:            resource.GetStatus().ProviderId.Id,
		ProviderError: apis.RandomString(),
	}, resource)
}

func AssertWorkflow[R pipelinesv1.Resource](
	newResource func() R,
	setUp func(resource R) base.Output,
	constructWorkflow func(string, R) (*argo.Workflow, error)) {

	testCtx := WorkflowTestHelper[R]{
		Resource: newResource(),
	}

	expectedOutput := setUp(testCtx.Resource)
	workflow, err := constructWorkflow("stub", testCtx.Resource)

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
	resource.SetNamespace("argo")
	resourceStatus := resource.GetStatus()
	resourceStatus.ProviderId.Provider = "stub"
	resource.SetStatus(resourceStatus)

	return resource
}
