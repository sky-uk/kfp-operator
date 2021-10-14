//go:build integration
// +build integration

package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/external"
	"github.com/walkerus/go-wiremock"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

const (
	TestTimeout = 120
)

var (
	k8sClient client.Client
	ctx       context.Context

	restCfg = rest.Config{
		Host:    "http://localhost:8080",
		APIPath: "/api",
	}

	pipelineSpec = pipelinesv1.PipelineSpec{
		Image:         "kfp-quickstart",
		TfxComponents: "pipeline.create_components",
	}

	wiremockClient *wiremock.Client
)

func TestPipelineControllerIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controller Suite")
}

var _ = BeforeSuite(func() {
	wiremockClient = wiremock.NewClient("http://localhost:8081")

	Expect(external.InitSchemes(scheme.Scheme)).To(Succeed())
	var err error
	k8sClient, err = client.New(&restCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	ctx = context.Background()
})

var _ = BeforeEach(func() {
	Expect(wiremockClient.Reset()).To(Succeed())

	Expect(wiremockClient.StubFor(wiremock.Get(wiremock.URLPathEqualTo("/apis/v1beta1/healthz")).
		WillReturn(
			`{"status": "ok"}`,
			map[string]string{"Content-Type": "application/json"},
			200,
		))).To(Succeed())
})
