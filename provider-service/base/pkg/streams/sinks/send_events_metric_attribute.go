package sinks

const sendEventsMetricResultKey = "result"

type sendEventsMetricAttribute int

const (
	Success sendEventsMetricAttribute = iota
	RecoverableFailure
	UnrecoverableFailure
	Discarded
)
