//go:build unit

package v1alpha6

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Context("Provider", func() {
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
