//go:build unit

package webhook

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebhookUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Unit Suite")
}
