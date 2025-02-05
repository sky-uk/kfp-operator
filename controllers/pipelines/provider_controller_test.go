//go:build unit

package pipelines

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Context("Provider Controller", func() {

	var (
		ctx                = context.Background()
		scheme             = runtime.NewScheme()
		provider           *pipelinesv1.Provider
		providerReconciler ProviderReconciler
	)

	err := pipelinesv1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	BeforeEach(func() {
		provider = pipelinesv1.RandomProvider()
		providerReconciler = ProviderReconciler{
			ProviderLoader: &testutil.MockProviderLoader{
				LoadProviderFunc: func() (pipelinesv1.Provider, error) {
					return *provider, nil
				},
			},
			DeploymentManager: &testutil.MockDeploymentManager{},
			ServiceManager:    &testutil.MockServiceManager{},
			StatusManager:     &testutil.MockStatusManager{},
		}
	})

	var _ = Describe("Reconcile", func() {

		Specify("Should Create a Deployment and Service if they do not currently exist", func() {
			deploymentCreateCalled := false
			serviceCreateCalled := false

			mockDeploymentMan := &testutil.MockDeploymentManager{
				GetFunc: func() (*appsv1.Deployment, error) {
					return nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "deployments"}, "deployment")
				},
				CreateFunc: func() error {
					deploymentCreateCalled = true
					return nil
				},
			}

			mockServiceMan := &testutil.MockServiceManager{
				GetFunc: func() (*v1.Service, error) {
					return nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "service"}, "service")
				},
				CreateFunc: func() error {
					serviceCreateCalled = true
					return nil
				},
			}

			providerReconciler.DeploymentManager = mockDeploymentMan
			providerReconciler.ServiceManager = mockServiceMan

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			Expect(deploymentCreateCalled).To(BeTrue())
			Expect(serviceCreateCalled).To(BeTrue())
		})
		Specify("Should Update a Deployment and Delete Service if they exist and provider resource is out of sync", func() {
			deploymentUpdateCalled := false
			serviceDeleteCalled := false

			mockDeploymentMan := &testutil.MockDeploymentManager{
				UpdateFunc: func() error {
					deploymentUpdateCalled = true
					return nil
				},
			}

			mockServiceMan := &testutil.MockServiceManager{
				DeleteFunc: func() error {
					serviceDeleteCalled = true
					return nil
				},
			}

			mockProviderLoader := &testutil.MockProviderLoader{
				LoadProviderFunc: func() (pipelinesv1.Provider, error) {
					outOfSyncProvider := *provider
					outOfSyncProvider.Generation = outOfSyncProvider.Status.ObservedGeneration + 1
					return outOfSyncProvider, nil
				},
			}

			providerReconciler.DeploymentManager = mockDeploymentMan
			providerReconciler.ServiceManager = mockServiceMan
			providerReconciler.ProviderLoader = mockProviderLoader

			result, err := providerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: provider.Namespace,
					Name:      provider.Name,
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			Expect(deploymentUpdateCalled).To(BeTrue())
			Expect(serviceDeleteCalled).To(BeTrue())
		})

	var _ = Describe("updateProviderStatus", func() {
		Specify("Should update the Provider status with the given conditions", func() {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			err := pipelinesv1.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&pipelinesv1.Provider{}).
				Build()

			pr := ProviderReconciler{
				ResourceReconciler: ResourceReconciler[*pipelinesv1.Provider]{
					EC: K8sExecutionContext{
						Client: controllers.OptInClient{
							Writer:       client,
							StatusClient: client,
							Cached:       client,
							NonCached:    client,
						},
						Scheme: scheme,
					},
					Config: config.KfpControllerConfigSpec{},
				},
				Scheme: scheme,
			}

			initialProvider := pipelinesv1.RandomProvider()
			expectedStartGeneration := int64(1)
			expectedEndGeneration := int64(2)

			initialProvider.Status.ObservedGeneration = expectedStartGeneration
			initialProvider.Generation = expectedEndGeneration

			err = client.Create(ctx, initialProvider)
			Expect(err).ToNot(HaveOccurred())

			err = client.Get(ctx, initialProvider.GetNamespacedName(), initialProvider)
			Expect(err).ToNot(HaveOccurred())
			Expect(initialProvider.Status.ObservedGeneration).To(Equal(expectedStartGeneration))

			err = pr.updateProviderStatus(ctx, initialProvider)
			Expect(err).ToNot(HaveOccurred())

			underTest := &pipelinesv1.Provider{}
			err = client.Get(ctx, initialProvider.GetNamespacedName(), underTest)
			Expect(err).ToNot(HaveOccurred())

			Expect(underTest.Status.ObservedGeneration).To(Equal(expectedEndGeneration))

		})
	})

	var _ = Describe("jsonToString", func() {
		Specify("Should return a plain string given a JSON string (no extra quotes or escape chars!)", func() {
			rawJson, err := json.Marshal("test")
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal("test"))
		})

		Specify("Should return a raw JSON string given a JSON object", func() {
			rawJson, err := json.Marshal(`{"key1": "value1", "key2": 42}`)
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal("{\"key1\": \"value1\", \"key2\": 42}"))
		})
	})
})
