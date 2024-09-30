//go:build integration

package webhook

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebhookIntegrationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Integration Suite")
}
