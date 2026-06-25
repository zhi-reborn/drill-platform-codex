package entity

import "testing"

func TestFlowCommandTableAndTerminalStatus(t *testing.T) {
	cmd := FlowCommand{}
	if got := cmd.TableName(); got != "drill_flow_command" {
		t.Fatalf("TableName() = %q, want %q", got, "drill_flow_command")
	}

	tests := []struct {
		name     string
		status   FlowCommandStatus
		terminal bool
	}{
		{name: "succeeded", status: FlowCommandSucceeded, terminal: true},
		{name: "failed", status: FlowCommandFailed, terminal: true},
		{name: "pending", status: FlowCommandPending, terminal: false},
		{name: "processing", status: FlowCommandProcessing, terminal: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.Status = tt.status
			if got := cmd.IsTerminal(); got != tt.terminal {
				t.Fatalf("IsTerminal() = %v, want %v", got, tt.terminal)
			}
		})
	}
}
