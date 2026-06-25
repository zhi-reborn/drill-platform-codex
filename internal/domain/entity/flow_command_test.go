package entity_test

import (
	"testing"

	"drill-platform/internal/domain/entity"
)

func TestFlowCommandTableName(t *testing.T) {
	cmd := entity.FlowCommand{}
	if got := cmd.TableName(); got != "drill_flow_command" {
		t.Fatalf("TableName() = %q, want %q", got, "drill_flow_command")
	}
}

func TestFlowCommandIsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status entity.FlowCommandStatus
		want   bool
	}{
		{name: "succeeded", status: entity.FlowCommandSucceeded, want: true},
		{name: "failed", status: entity.FlowCommandFailed, want: true},
		{name: "pending", status: entity.FlowCommandPending, want: false},
		{name: "processing", status: entity.FlowCommandProcessing, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := entity.FlowCommand{Status: tt.status}
			if got := cmd.IsTerminal(); got != tt.want {
				t.Fatalf("IsTerminal() = %t, want %t", got, tt.want)
			}
		})
	}
}
