//go:build unit

package base

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProvidersUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provider base")
}
