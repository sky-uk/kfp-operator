package v1alpha5

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
)

var PropertyBased = MustPassRepeatedly(5)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

var syncStateComparer = cmp.FilterPath(
	func(p cmp.Path) bool {
		return p.Last().String() == ".SynchronizationState"
	},
	cmp.Comparer(
		func(lhs, rhs apis.SynchronizationState) bool {
			return lhs == rhs || (lhs == "" && rhs == "Unknown") || (lhs == "Unknown" && rhs == "")
		},
	),
)
