//go:build unit

package server

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServerUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Unit Suite")
}
