//go:build unit

package publisher

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPublisherUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Publisher Unit Suite")
}
