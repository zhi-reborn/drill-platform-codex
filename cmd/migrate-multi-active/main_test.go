package main

import (
	"os"
	"strings"
	"testing"
)

func TestLogStepColumnPrefersRenamedColumn(t *testing.T) {
	if got := logStepColumn(true); got != "task_instance_id" {
		t.Fatalf("logStepColumn(true) = %q, want task_instance_id", got)
	}
	if got := logStepColumn(false); got != "step_instance_id" {
		t.Fatalf("logStepColumn(false) = %q, want step_instance_id", got)
	}
}

func TestVerifyRequirementsCoverMultiActiveSchema(t *testing.T) {
	reqs := verifyRequirements()
	want := []string{
		"table:drill_flow_command",
		"table:drill_worker_epoch",
		"column:drill_flow_command.worker_epoch",
		"column:drill_flow_command.lease_token",
		"column:drill_flow_command.attempt_count",
		"column:drill_instance_step_log.command_id",
		"column:notification.command_id",
	}
	for _, key := range want {
		if _, ok := reqs[key]; !ok {
			t.Fatalf("verifyRequirements missing %s; got %#v", key, reqs)
		}
	}
}

func TestPrimarySQLMigrationAvoidsStoredProcedures(t *testing.T) {
	data, err := os.ReadFile("../../scripts/migration/2026-07-05-migrate-to-multi-active.sql")
	if err != nil {
		t.Fatalf("read migration sql: %v", err)
	}
	sql := strings.ToUpper(string(data))
	for _, forbidden := range []string{"CREATE PROCEDURE", "DROP PROCEDURE", "CALL "} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("primary migration contains %q, which is not TiDB compatible", forbidden)
		}
	}
}
