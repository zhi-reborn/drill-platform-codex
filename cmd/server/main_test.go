package main

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApplyGinModeUsesReleaseFromConfig(t *testing.T) {
	t.Cleanup(func() {
		gin.SetMode(gin.DebugMode)
	})

	gin.SetMode(gin.DebugMode)
	applyGinMode("release")

	if got := gin.Mode(); got != gin.ReleaseMode {
		t.Fatalf("gin mode = %q, want %q", got, gin.ReleaseMode)
	}
}
