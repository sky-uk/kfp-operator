//go:build unit

package run_completion_event_trigger

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunCompletionEventTrigger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Completion Event Trigger Unit Suite")
}
