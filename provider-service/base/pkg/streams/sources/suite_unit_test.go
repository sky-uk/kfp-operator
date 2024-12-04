//go:build unit

package sources

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStreamsUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Streams Unit Suite")
}
