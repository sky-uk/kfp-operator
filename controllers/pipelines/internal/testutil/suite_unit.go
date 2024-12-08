//go:build unit

package testutil

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var PropertyBased = MustPassRepeatedly(5)

func TestPipelineControllersUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers Unit Suite")
}
