package apis

type SynchronizationState string

const (
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
	Unknown   SynchronizationState = "Unknown"
)

var validStates = map[string]SynchronizationState{
	string(Creating):  Creating,
	string(Succeeded): Succeeded,
	string(Updating):  Updating,
	string(Deleting):  Deleting,
	string(Deleted):   Deleted,
	string(Failed):    Failed,
}

func SynchronisationState(s string) SynchronizationState {
	state, ok := validStates[s]
	if !ok {
		state = Unknown
	}
	return state
}

const Group = "pipelines.kubeflow.org"

// +kubebuilder:object:generate=true
type NamedValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (nv NamedValue) GetKey() string {
	return nv.Name
}

func (nv NamedValue) GetValue() string {
	return nv.Value
}
