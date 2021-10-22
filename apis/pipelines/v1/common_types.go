package v1

type SynchronizationState string

const (
	Unknown   SynchronizationState = ""
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
)

type Status struct {
	KfpId                string               `json:"kfpId,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string               `json:"version,omitempty"`
}
