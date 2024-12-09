//go:build decoupled

package pipelines

import (
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceTestHelper[R pipelinesv1.Resource] struct {
	WorkflowTestHelper[R]
	Resource R
}

func Create[R pipelinesv1.Resource](resource R) ResourceTestHelper[R] {
	K8sClient.Create(Ctx, resource)

	return ResourceTestHelper[R]{
		Resource: resource,
		WorkflowTestHelper: WorkflowTestHelper[R]{
			Resource: resource,
		},
	}
}

func (testCtx ResourceTestHelper[R]) ToMatch(matcher func(Gomega, R)) func(Gomega) {
	return func(g Gomega) {
		Expect(K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)).To(Succeed())
		matcher(g, testCtx.Resource)
	}
}

func (testCtx ResourceTestHelper[R]) Exists() error {
	return K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) Update(updateFunc func(R)) error {
	if err := K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	updateFunc(testCtx.Resource)

	return K8sClient.Update(Ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) UpdateStatus(updateFunc func(R)) error {
	if err := K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	updateFunc(testCtx.Resource)

	return K8sClient.Status().Update(Ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) Delete() error {
	if err := K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource); err != nil {
		return err
	}

	return K8sClient.Delete(Ctx, testCtx.Resource)
}

func (testCtx ResourceTestHelper[R]) EmittedEventsToMatch(matcher func(Gomega, []v1.Event)) func(Gomega) {
	return func(g Gomega) {
		eventList := &v1.EventList{}
		Expect(
			K8sClient.List(
				Ctx,
				eventList,
				client.MatchingFields{"involvedObject.name": testCtx.Resource.GetName()},
			),
		).To(Succeed())

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
	Expect(K8sClient.Get(Ctx, testCtx.Resource.GetNamespacedName(), testCtx.Resource)).To(Succeed())

	testCtx.Resource.SetStatus(pipelinesv1.Status{
		SynchronizationState: apis.Succeeded,
		Version:              testCtx.Resource.ComputeVersion(),
		Provider: pipelinesv1.ProviderAndId{
			Name: Provider.Name,
			Id:   apis.RandomString(),
		},
	})

	Expect(K8sClient.Status().Update(Ctx, testCtx.Resource)).To(Succeed())

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
