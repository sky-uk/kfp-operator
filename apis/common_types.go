package apis

import (
	"context"
	"encoding/json"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SynchronizationState string

const (
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
)

const Group = "pipelines.kubeflow.org"

//+kubebuilder:object:generate=true
type NamedValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//+kubebuilder:object:generate=true
type Status struct {
	KfpId                string               `json:"kfpId,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string               `json:"version,omitempty"`
	ObservedGeneration   int64                `json:"observedGeneration,omitempty"`
}

var Annotations = struct {
	Debug string
}{
	Debug: Group + "/debug",
}

type DebugOptions struct {
	KeepWorkflows bool `json:"keepWorkflows,omitempty"`
}

func (options DebugOptions) WithDefaults(defaults DebugOptions) DebugOptions {
	return DebugOptions{
		KeepWorkflows: options.KeepWorkflows || defaults.KeepWorkflows,
	}
}

func DebugOptionsFromAnnotations(ctx context.Context, annotations map[string]string) DebugOptions {
	logger := log.FromContext(ctx)
	debugOptions := DebugOptions{}

	if debugAnnotation, ok := annotations[Annotations.Debug]; ok {
		if err := json.Unmarshal([]byte(debugAnnotation), &debugOptions); err != nil {
			logger.Error(err, "error unmarshalling pipeline annotations")
		}
	}

	return debugOptions
}

func AnnotationsFromDebugOptions(ctx context.Context, debugOptions DebugOptions) map[string]string {
	logger := log.FromContext(ctx)

	if debugAnnotation, err := json.Marshal(debugOptions); err != nil {
		logger.Error(err, "error marshalling debug options into json")
		return map[string]string{}
	} else {
		return map[string]string{
			Annotations.Debug: string(debugAnnotation),
		}
	}
}
