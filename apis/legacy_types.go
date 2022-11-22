package apis

var Annotations = struct {
	Debug string
}{
	Debug: Group + "/debug",
}

type DebugOptions struct {
	KeepWorkflows bool `json:"keepWorkflows,omitempty"`
}
