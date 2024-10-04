//go:build unit

package nats_event_trigger

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNATSEventTrigger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NATs Event Trigger Unit Suite")
}
