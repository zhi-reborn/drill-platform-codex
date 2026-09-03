package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type ActivityRecorder interface {
	Touch(context.Context, uint64) error
}

func RecordPresence(recorder ActivityRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID := GetUserID(c); userID != 0 {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 250*time.Millisecond)
			// Presence is best-effort; Redis failure must not reject the business request.
			_ = recorder.Touch(ctx, userID)
			cancel()
		}
		c.Next()
	}
}
