package apis

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

var ResourceAnnotations = struct {
	Provider string
}{
	Provider: Group + "/provider",
}
