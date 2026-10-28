//go:build unit

package pipelines

import (
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReplacementScheduleSafety(t *testing.T) {
	original := pipelineshub.RunSchedule{}
	replacement := pipelineshub.RunSchedule{}
	original.Spec.Schedule = pipelineshub.RandomSchedule()
	replacement.Spec.Schedule = pipelineshub.RandomSchedule()
	if compareRunSchedules(original, replacement) {
		t.Skip("random schedules unexpectedly identical")
	}
	actual := []pipelineshub.RunSchedule{original}
	desired := []pipelineshub.RunSchedule{replacement}
	if allDesiredSchedulesSucceeded(desired, actual) {
		t.Fatal("old schedule must not count as successful replacement")
	}
	state, _ := aggregateDesiredScheduleState(desired, actual)
	if state != apis.Updating {
		t.Fatalf("expected Updating, got %v", state)
	}
	replacement.Status.Conditions = []metav1.Condition{{Type: apis.ConditionTypes.SynchronizationSucceeded, Reason: string(apis.Failed), Message: "provider rejected replacement"}}
	actual = append(actual, replacement)
	if allDesiredSchedulesSucceeded(desired, actual) {
		t.Fatal("failed replacement must not permit deletion")
	}
	state, msg := aggregateDesiredScheduleState(desired, actual)
	if state != apis.Failed || msg != "provider rejected replacement" {
		t.Fatalf("expected provider failure, got %v: %s", state, msg)
	}
	actual[1].Status.Conditions[0].Reason = string(apis.Succeeded)
	if !allDesiredSchedulesSucceeded(desired, actual) {
		t.Fatal("successful replacement should permit deletion")
	}
	if !allDesiredSchedulesSucceeded(nil, actual) {
		t.Fatal("intentional removal of all schedules should remain permitted")
	}
}
