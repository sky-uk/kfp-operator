//go:build unit

package sinks

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSinksUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sinks Unit Suite")
}
