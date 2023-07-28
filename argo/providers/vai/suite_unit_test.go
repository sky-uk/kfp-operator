//go:build unit

package vai

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VAI Provider Unit Suite")
}
