//go:build decoupled

package pipelines

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
)

func HaveReason(reason string) types.GomegaMatcher {
	return &EventHasReasonMatcher{
		Reason: reason,
	}
}

type EventHasReasonMatcher struct {
	Reason string
}

func (matcher *EventHasReasonMatcher) Match(actual interface{}) (success bool, err error) {
	event, ok := actual.(v1.Event)
	if !ok {
		return false, fmt.Errorf("EventHasReasonMatcher matcher expects a v1.Event.  Got:\n%s", format.Object(actual, 1))
	}

	return event.Reason == matcher.Reason, nil
}

func (matcher *EventHasReasonMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have reason", matcher.Reason)
}

func (matcher *EventHasReasonMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have reason", matcher.Reason)
}
