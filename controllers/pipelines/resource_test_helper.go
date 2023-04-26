//go:build decoupled
// +build decoupled

package pipelines

import (
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceTestHelper[R pipelinesv1.Resource] struct {
	WorkflowTestHelper[R]
	Resource R
}

func Create[R pipelinesv1.Resource](resource R) ResourceTestHelper[R] {
	k8sClient.Create(ctx, resource)

	return ResourceTestHelper[R]{
		Resource: resource,
		WorkflowTestHelper: WorkflowTestHelper[R]{
			Resource: resource,
		},
	}
}

func (testCtx ResourceTestHelper[R]) ToMatch(matcher func(Gomega, R)) func(Gomega) {
	return func(g Gomega) {
		Expect(k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)).To(Succeed())
		matcher(g, testCtx.Resource)
	}
}

func (testCtx ResourceTestHelper[R]) Exists() error {
	return k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) Update(updateFunc func(R)) error {
	if err := k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	updateFunc(testCtx.Resource)

	return k8sClient.Update(ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) UpdateStatus(updateFunc func(R)) error {
	if err := k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	updateFunc(testCtx.Resource)

	return k8sClient.Status().Update(ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) Delete() error {
	if err := k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	return k8sClient.Delete(ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(k8sClient.List(ctx, eventList, client.MatchingFields{"involvedObject.name": testCtx.Resource.GetName()})).To(Succeed())

		matcher(g, eventList.Items)
	}
}

func (testCtx ResourceTestHelper[R]) UpdateStable(updateFunc func(resource R)) {
	Expect(testCtx.Update(updateFunc)).To(Succeed())

	Eventually(testCtx.ToMatch(func(g Gomega, resource R) {
		g.Expect(resource.GetStatus().SynchronizationState).To(Equal(apis.Updating))
	})).Should(Succeed())

	testCtx.UpdateToSucceeded()
}

func (testCtx ResourceTestHelper[R]) UpdateToSucceeded() {
	Expect(k8sClient.Get(ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)).To(Succeed())

	testCtx.Resource.SetStatus(pipelinesv1.Status{
		SynchronizationState: apis.Succeeded,
		Version:              testCtx.Resource.ComputeVersion(),
		ProviderId: pipelinesv1.ProviderAndId{
			Provider: testConfig.DefaultProvider,
			Id:       apis.RandomString(),
		},
	})

	Expect(k8sClient.Status().Update(ctx, testCtx.Resource)).To(Succeed())

	Eventually(testCtx.ToMatch(func(g Gomega, resource R) {
		g.Expect(resource.GetGeneration()).To(Equal(resource.GetStatus().ObservedGeneration))
	})).Should(Succeed())
}

func CreateSucceeded[R pipelinesv1.Resource](resource R) ResourceTestHelper[R] {
	testCtx := CreateStable(resource)
	testCtx.UpdateToSucceeded()

	return testCtx
}

func CreateStable[R pipelinesv1.Resource](resource R) ResourceTestHelper[R] {
	testCtx := Create(resource)
	Eventually(testCtx.ToMatch(func(g Gomega, resource R) {
		g.Expect(resource.GetStatus().SynchronizationState).To(Equal(apis.Creating))
	})).Should(Succeed())

	return testCtx
}
