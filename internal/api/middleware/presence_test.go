package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type recordingPresence struct {
	ids     []uint64
	err     error
	bounded bool
}

func (p *recordingPresence) Touch(ctx context.Context, id uint64) error {
	p.ids = append(p.ids, id)
	_, p.bounded = ctx.Deadline()
	return p.err
}

func TestRecordPresenceIgnoresAnonymousAndKeepsBusinessSuccess(t *testing.T) {
	for _, id := range []uint64{0, 7} {
		p := &recordingPresence{err: errors.New("Redis down")}
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(CtxUserIDInt, id) })
		r.Use(RecordPresence(p))
		r.GET("/test", func(c *gin.Context) { c.Status(200) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))
		if w.Code != 200 {
			t.Fatalf("business status = %d", w.Code)
		}
		if id == 0 && len(p.ids) != 0 {
			t.Fatal("anonymous request counted")
		}
		if id != 0 && (len(p.ids) != 1 || p.ids[0] != id || !p.bounded) {
			t.Fatalf("missing bounded recording: %+v", p)
		}
	}
}

func TestRecordPresenceNotReachedWithInvalidToken(t *testing.T) {
	p := &recordingPresence{}
	r := gin.New()
	r.Use(JWTAuth(JWTConfig{Secret: "test-secret"}), RecordPresence(p))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/test?token=invalid", nil))
	if w.Code != 401 || len(p.ids) != 0 {
		t.Fatalf("invalid auth = %d, ids=%v", w.Code, p.ids)
	}
}
