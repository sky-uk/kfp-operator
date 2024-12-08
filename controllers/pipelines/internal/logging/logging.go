package logging

var Keys = struct {
	Workflow  string
	Command   string
	Duration  string
	Status    string
	OldStatus string
	NewStatus string
}{
	Workflow:  "workflow",
	Command:   "command",
	Duration:  "duration",
	Status:    "status",
	OldStatus: "oldStatus",
	NewStatus: "newStatus",
}
