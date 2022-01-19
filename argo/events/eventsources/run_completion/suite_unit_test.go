//go:build unit
// +build unit

package run_completion

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run completion Unit Suite")
}
