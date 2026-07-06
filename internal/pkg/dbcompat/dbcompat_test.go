package dbcompat

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsMySQLDialect(t *testing.T) {
	if !IsMySQLDialect("mysql") {
		t.Fatal("mysql dialect should be treated as MySQL-compatible")
	}
	if IsMySQLDialect("sqlite") {
		t.Fatal("sqlite dialect should not be treated as MySQL-compatible")
	}
}

func TestIsTiDBNoopFunctionError(t *testing.T) {
	err := &mysql.MySQLError{
		Number:  1235,
		Message: "function get_lock has only noop implementation in tidb now, use tidb_enable_noop_functions to enable these functions",
	}

	if !IsTiDBNoopFunctionError(err, "get_lock") {
		t.Fatal("expected TiDB noop GET_LOCK error")
	}
	if IsTiDBNoopFunctionError(err, "release_lock") {
		t.Fatal("GET_LOCK error must not match RELEASE_LOCK")
	}
	if IsTiDBNoopFunctionError(errors.New("function get_lock has only noop implementation in tidb now"), "get_lock") {
		t.Fatal("non-MySQL errors must not be treated as TiDB noop errors")
	}
}

func TestIgnoreTiDBNoopFunctionError(t *testing.T) {
	err := &mysql.MySQLError{
		Number:  1235,
		Message: "function release_lock has only noop implementation in tidb now",
	}

	if got := IgnoreTiDBNoopFunctionError(err, "release_lock"); got != nil {
		t.Fatalf("expected TiDB noop error ignored, got %v", got)
	}
	if got := IgnoreTiDBNoopFunctionError(err, "get_lock"); got == nil {
		t.Fatal("mismatched function error must be preserved")
	}
}
