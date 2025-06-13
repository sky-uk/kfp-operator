package apis

import (
	"sigs.k8s.io/controller-runtime/pkg/config"
	"time"
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

// +kubebuilder:object:generate=false
type JsonPatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

// +kubebuilder:object:generate=true
type ControllerWrapper struct {
	GroupKindConcurrency    map[string]int `json:"groupKindConcurrency,omitempty"`
	MaxConcurrentReconciles int            `json:"maxConcurrentReconciles,omitempty"`
	CacheSyncTimeout        time.Duration  `json:"cacheSyncTimeout,omitempty"`
	RecoverPanic            *bool          `json:"recoverPanic,omitempty"`
	NeedLeaderElection      *bool          `json:"needLeaderElection,omitempty"`
}

func (cw *ControllerWrapper) FromController(c config.Controller) {
	cw.GroupKindConcurrency = c.GroupKindConcurrency
	cw.MaxConcurrentReconciles = c.MaxConcurrentReconciles
	cw.CacheSyncTimeout = c.CacheSyncTimeout
	cw.RecoverPanic = c.RecoverPanic
	cw.NeedLeaderElection = c.NeedLeaderElection
}

func (cw *ControllerWrapper) ToController() config.Controller {
	return config.Controller{
		GroupKindConcurrency:    cw.GroupKindConcurrency,
		MaxConcurrentReconciles: cw.MaxConcurrentReconciles,
		CacheSyncTimeout:        cw.CacheSyncTimeout,
		RecoverPanic:            cw.RecoverPanic,
		NeedLeaderElection:      cw.NeedLeaderElection,
	}
}
