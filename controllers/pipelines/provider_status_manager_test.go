//go:build unit

package pipelines

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Context("Provider Status Manager", func() {

	var (
		ctx           = context.Background()
		provider      *pipelineshub.Provider
		client        k8sClient.Client
		statusManager StatusManager
	)

	BeforeEach(func() {
		provider = pipelineshub.RandomProvider()
		client = fake.NewClientBuilder().
			WithScheme(testutil.SchemeWithCrds()).
			WithStatusSubresource(&pipelineshub.Provider{}).
			Build()

		optInClient := &controllers.OptInClient{
			Writer:       client,
			StatusClient: client,
			Cached:       client,
			NonCached:    client,
		}

		statusManager = StatusManager{
			client: optInClient,
		}
	})

	var _ = Describe("UpdateProviderStatus", func() {

		It("should not error if the provider status is updated successfully", func() {
			provider.Status = pipelineshub.Status{}
			provider.Generation = 1

			err := client.Create(ctx, provider)
			Expect(err).ToNot(HaveOccurred())

			expectedMessage := "test"
			expectedStatus := apis.Succeeded
			err = statusManager.UpdateProviderStatus(ctx, provider, expectedStatus, expectedMessage)
			Expect(err).ToNot(HaveOccurred())

			err = client.Get(ctx, k8sClient.ObjectKey{Name: provider.Name, Namespace: provider.Namespace}, provider)
			Expect(err).ToNot(HaveOccurred())

			Expect(provider.Status.Conditions[0].Message).To(Equal(expectedMessage))
			Expect(provider.Status.Conditions.SynchronizationSucceeded().Reason).To(Equal(string(expectedStatus)))
			Expect(provider.Status.ObservedGeneration).To(Equal(provider.Generation))
		})

		It("should not update observed generation if the provider status update is failed", func() {
			provider.Status = pipelineshub.Status{}
			provider.Generation = 1

			err := client.Create(ctx, provider)
			Expect(err).ToNot(HaveOccurred())

			expectedMessage := "test"
			expectedStatus := apis.Failed
			err = statusManager.UpdateProviderStatus(ctx, provider, expectedStatus, expectedMessage)
			Expect(err).ToNot(HaveOccurred())

			err = client.Get(ctx, k8sClient.ObjectKey{Name: provider.Name, Namespace: provider.Namespace}, provider)
			Expect(err).ToNot(HaveOccurred())

			Expect(provider.Status.Conditions[0].Message).To(Equal(expectedMessage))
			Expect(provider.Status.Conditions.SynchronizationSucceeded().Reason).To(Equal(string(expectedStatus)))
			Expect(provider.Status.ObservedGeneration).ToNot(Equal(provider.Generation))
		})

		It("should return error if the status is not updated", func() {
			// No provider is created so no status to update
			err := statusManager.UpdateProviderStatus(ctx, provider, apis.Succeeded, "test")
			Expect(err).To(HaveOccurred())
		})
	})
})
