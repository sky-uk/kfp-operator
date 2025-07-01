//go:build unit

package provider

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKfpProviderServicesUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kfp Provider Service")
}
