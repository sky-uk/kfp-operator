//go:build unit

package pipelines

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Context("Provider Controller", func() {

	var (
		ctx                = context.Background()
		scheme             = runtime.NewScheme()
		deployment         = &appsv1.Deployment{}
		service            = &v1.Service{}
		providerReconciler ProviderReconciler
		provider           *pipelinesv1.Provider
		mockProviderLoader *testutil.MockProviderLoader
		mockDeploymentMan  *testutil.MockDeploymentManager
		mockServiceMan     *testutil.MockServiceManager
		mockStatusMan      *testutil.MockStatusManager
	)

	err := pipelinesv1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	BeforeEach(func() {
		provider = pipelinesv1.RandomProvider()
		provider.Generation = 1
		provider.Status.ObservedGeneration = provider.Generation

		mockProviderLoader = &testutil.MockProviderLoader{}
		mockDeploymentMan = &testutil.MockDeploymentManager{}
		mockServiceMan = &testutil.MockServiceManager{}
		mockStatusMan = &testutil.MockStatusManager{}

		providerReconciler = ProviderReconciler{
			ProviderLoader:    mockProviderLoader,
			DeploymentManager: mockDeploymentMan,
			ServiceManager:    mockServiceMan,
			StatusManager:     mockStatusMan,
		}
	})

	var _ = Describe("Reconcile", func() {

		Specify("Should Create a Deployment and Service if they do not currently exist", func() {
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "deployments"}, "deployment")).
				On("Construct", provider).Return(deployment, nil).
				On("Create", deployment, provider).Return(nil)

			mockServiceMan.
				On("Get", provider).Return(nil, errors.NewNotFound(schema.GroupResource{Group: "", Resource: "services"}, "service")).
				On("Construct", provider).Return(service).
				On("Create", service, provider).Return(nil)

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			mockDeploymentMan.AssertCalled(GinkgoT(), "Create", deployment, provider)
			mockServiceMan.AssertCalled(GinkgoT(), "Create", service, provider)

			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Creating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Succeeded, "")

		})

		Specify("Should Update a Deployment and Delete Service if they exist and provider resource is out of sync", func() {
			//Out of sync provider
			provider.Generation = provider.Status.ObservedGeneration + 1
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(deployment, nil).
				On("Construct", provider).Return(deployment, nil).
				On("Equal", deployment, deployment).Return(true).
				On("Update", deployment, deployment, provider).Return(nil)

			mockServiceMan.
				On("Get", provider).Return(service, nil).
				On("Construct", provider).Return(service).
				On("Equal", service, service).Return(true).
				On("Delete", service).Return(nil)

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			mockDeploymentMan.AssertCalled(GinkgoT(), "Update", deployment, deployment, provider)
			mockServiceMan.AssertCalled(GinkgoT(), "Delete", service)

			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Updating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Succeeded, "")
		})

		Specify("Should update a deployment if it is out of sync", func() {
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(deployment, nil).
				On("Construct", provider).Return(deployment, nil).
				On("Equal", deployment, deployment).Return(false).
				On("Update", deployment, deployment, provider).Return(nil)

			mockServiceMan.
				On("Get", provider).Return(service, nil).
				On("Construct", provider).Return(service).
				On("Equal", service, service).Return(true)

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			mockDeploymentMan.AssertCalled(GinkgoT(), "Update", deployment, deployment, provider)
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Updating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Succeeded, "")
		})

		Specify("Should delete a service if it is out of sync", func() {
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(deployment, nil).
				On("Construct", provider).Return(deployment, nil).
				On("Equal", deployment, deployment).Return(true)

			mockServiceMan.
				On("Get", provider).Return(service, nil).
				On("Construct", provider).Return(service).
				On("Equal", service, service).Return(false).
				On("Delete", service).Return(nil)

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			mockServiceMan.AssertCalled(GinkgoT(), "Delete", service)

			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Updating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Succeeded, "")
		})

		Specify("Should Update a deployment if it is out of sync", func() {
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(deployment, nil).
				On("Construct", provider).Return(deployment, nil).
				On("Equal", deployment, deployment).Return(false).
				On("Update", deployment, deployment, provider).Return(nil)

			mockServiceMan.
				On("Get", provider).Return(service, nil).
				On("Construct", provider).Return(service).
				On("Equal", service, service).Return(true)

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			mockDeploymentMan.AssertCalled(GinkgoT(), "Update", deployment, deployment, provider)
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Updating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Succeeded, "")
		})

		Specify("Should set state as failed when error occurs", func() {
			mockProviderLoader.On("LoadProvider", mock.Anything, mock.Anything).Return(*provider, nil)

			mockDeploymentMan.
				On("Get", provider).Return(deployment, nil).
				On("Construct", provider).Return(deployment, nil).
				On("Equal", deployment, deployment).Return(false).
				On("Update", deployment, deployment, provider).Return(errors.NewBadRequest("error"))

			mockStatusMan.On("UpdateProviderStatus", provider, mock.Anything, mock.Anything).Return(nil)

			_, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})

			Expect(err).To(HaveOccurred())

			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Updating, "")
			mockStatusMan.AssertCalled(GinkgoT(), "UpdateProviderStatus", provider, apis.Failed, "Failed to reconcile subresource deployment")
		})
	})
})
