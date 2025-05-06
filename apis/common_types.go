package apis

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

// +kubebuilder:object:generate=false
type JsonPatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}
