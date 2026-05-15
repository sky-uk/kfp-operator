package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	runtimeMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	labelStartedBy                  = "started_by"
	labelStartedByResource          = "started_by_resource"
	labelStartedByResourceNamespace = "started_by_resource_namespace"
	labelRunConfiguration           = "run_configuration"
	labelNamespace                  = "namespace"
	labelPipeline                   = "pipeline"
)

type RunTriggeredLabels struct {
	StartedBy                  string
	StartedByResource          string
	StartedByResourceNamespace string
	RunConfiguration           string
	Namespace                  string
	Pipeline                   string
}

func (rtl RunTriggeredLabels) ToPrometheusLabels() prometheus.Labels {
	return prometheus.Labels{
		labelStartedBy:                  rtl.StartedBy,
		labelStartedByResource:          rtl.StartedByResource,
		labelStartedByResourceNamespace: rtl.StartedByResourceNamespace,
		labelRunConfiguration:           rtl.RunConfiguration,
		labelNamespace:                  rtl.Namespace,
		labelPipeline:                   rtl.Pipeline,
	}
}

var RunsTriggeredTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kfp_operator_onchange_runs_total",
		Help: "Total number of runs triggered by RunConfiguration onChange events.",
	},
	[]string{labelStartedBy, labelStartedByResource, labelStartedByResourceNamespace, labelRunConfiguration, labelNamespace, labelPipeline},
)

func init() {
	runtimeMetrics.Registry.MustRegister(RunsTriggeredTotal)
}