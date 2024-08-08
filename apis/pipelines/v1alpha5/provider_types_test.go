//go:build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Provider", func() {
	var _ = Describe("SetSynchronizationState", func() {

		Specify("Should update the state of a given provider", func() {
			var allStates = []apis.SynchronizationState{
				apis.Creating,
				apis.Succeeded,
				apis.Updating,
				apis.Deleting,
				apis.Deleted,
				apis.Failed,
			}

			for _, state := range allStates {
				var provider = Provider{}
				var generation = common.RandomInt64()
				var message = common.RandomString()

				provider.Status.SetSynchronizationState(state, message, generation)

				var result = provider.Status.Conditions[0]
				Expect(result.ObservedGeneration).To(Equal(generation))
				Expect(result.Reason).To(Equal(string(state)))
				Expect(result.Message).To(Equal(message))
			}
		})
	})

	var _ = Describe("GetNamespacedName", func() {
		Specify("Should return a namespaced name for the resource", func() {
			var name = common.RandomString()
			var namespace = common.RandomString()
			var provider = Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
			Expect(provider.GetNamespacedName()).Should(Equal(types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}))
		})
	})
})
