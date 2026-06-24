package entity_test

import (
	"testing"

	"drill-platform/internal/domain/entity"
)

func TestFlowCommandTableAndTerminalStatus(t *testing.T) {
	cmd := entity.FlowCommand{Status: entity.FlowCommandSucceeded}
	if got := cmd.TableName(); got != "drill_flow_command" {
		t.Fatalf("TableName() = %q, want %q", got, "drill_flow_command")
	}
	if !cmd.IsTerminal() {
		t.Fatal("succeeded command should be terminal")
	}
}
