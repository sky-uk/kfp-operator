//go:build unused
// +build unused

/**
This file exists to preserve imports that are not referenced in code from being removed by `go mod tidy`
 */
package events

import (
	argo_events "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
)

var something = argo_events.Sensor{}
