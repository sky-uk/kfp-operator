//go:build unit

package internal

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VAI Provider Unit Suite")
}
