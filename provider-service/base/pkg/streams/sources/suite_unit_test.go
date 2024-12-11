//go:build unit

package sources

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSourcesUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sources Unit Suite")
}
