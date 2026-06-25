package events

import (
	"context"
	"encoding/json"
	"time"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	DrillID   uint64          `json:"drill_id,omitempty"`
	UserID    uint64          `json:"user_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type Subscriber interface {
	Subscribe(context.Context, func(Event)) error
	Healthy() bool
}
