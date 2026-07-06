package repository

import (
	"testing"

	"gorm.io/gorm/logger"
)

func TestGormLogLevel(t *testing.T) {
	if got := gormLogLevel(&Config{}); got != logger.Silent {
		t.Fatalf("default log level = %v, want %v", got, logger.Silent)
	}
	if got := gormLogLevel(&Config{LogSQL: true}); got != logger.Info {
		t.Fatalf("enabled log level = %v, want %v", got, logger.Info)
	}
}
