package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CurrentVersion is the schema version carried by every WSMessage envelope.
// Bumping this allows consumers to reject incompatible payloads.
const CurrentVersion = 1

// WSMessage is the canonical, versioned event envelope shared by Redis
// pub/sub and WebSocket delivery. Every realtime event — whether it flows
// through Redis (cross-node fanout) or directly to a WebSocket client — uses
// this single shape so that producers, the bus, and consumers agree on the
// contract.
//
// JSON shape:
//
//	{"id":"event-id","version":1,"type":"step.updated","drill_id":72,
//	 "occurred_at":"2026-06-25T12:00:00Z","data":{}}
type WSMessage struct {
	ID         string          `json:"id"`
	Version    int             `json:"version"`
	Type       string          `json:"type"`
	DrillID    uint64          `json:"drill_id,omitempty"`
	UserID     uint64          `json:"user_id,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

// Publisher publishes a WSMessage onto the realtime bus (Redis). Producers
// (services) call Publish; they never write to the local WebSocket hub
// directly — the subscriber owns local fanout.
type Publisher interface {
	Publish(context.Context, WSMessage) error
}

// Subscriber consumes WSMessages from the realtime bus and invokes the
// handler for each deduped event. Healthy reports whether the subscription
// is currently active; /ready gates on this so a node that has lost its
// Redis subscription drains traffic.
type Subscriber interface {
	Subscribe(context.Context, func(WSMessage)) error
	Healthy() bool
}

// NewWSMessage constructs a canonical envelope with a generated ID, the
// current schema version, and OccurredAt set to now. Callers supply the
// Type, targeting (DrillID and/or UserID), and Data payload. Using this
// helper guarantees every event on the bus carries a unique ID (required
// for subscriber dedup) and the current version.
func NewWSMessage(eventType string, drillID, userID uint64, data json.RawMessage) WSMessage {
	return WSMessage{
		ID:         uuid.NewString(),
		Version:    CurrentVersion,
		Type:       eventType,
		DrillID:    drillID,
		UserID:     userID,
		OccurredAt: time.Now().UTC(),
		Data:       data,
	}
}
