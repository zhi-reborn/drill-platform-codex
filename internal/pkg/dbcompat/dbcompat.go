package dbcompat

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func IsMySQLDialect(dialect string) bool {
	return dialect == "mysql"
}

func IsTiDBNoopFunctionError(err error, functionName string) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1235 {
		return false
	}
	message := strings.ToLower(mysqlErr.Message)
	return strings.Contains(message, "tidb") &&
		strings.Contains(message, "noop implementation") &&
		strings.Contains(message, strings.ToLower(functionName))
}

func IgnoreTiDBNoopFunctionError(err error, functionName string) error {
	if IsTiDBNoopFunctionError(err, functionName) {
		return nil
	}
	return err
}
