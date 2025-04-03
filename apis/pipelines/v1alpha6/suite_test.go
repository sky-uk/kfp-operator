package v1alpha6

import (
	"github.com/google/go-cmp/cmp"
	"github.com/sky-uk/kfp-operator/apis"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
