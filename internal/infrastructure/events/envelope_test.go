package events

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// TestPublisherEmitsCanonicalEnvelope verifies that Publish serializes a
// WSMessage with the versioned canonical envelope fields (id, version, type,
// drill_id, occurred_at, data) — the single shape shared by Redis and
// WebSocket.
func TestPublisherEmitsCanonicalEnvelope(t *testing.T) {
	client := &fakeRedisClient{}
	bus := NewRedisBus(client)

	occurred := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	msg := WSMessage{
		ID:         "evt-canonical",
		Version:    CurrentVersion,
		Type:       "step.updated",
		DrillID:    72,
		OccurredAt: occurred,
		Data:       json.RawMessage(`{"status":"completed"}`),
	}

	if err := bus.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if client.channel != EventsChannel {
		t.Fatalf("Publish() channel = %q, want %q", client.channel, EventsChannel)
	}

	payload, ok := client.message.(string)
	if !ok {
		t.Fatalf("Publish() message type = %T, want string", client.message)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		t.Fatalf("Publish() payload is not valid JSON: %v", err)
	}

	required := []string{"id", "version", "type", "drill_id", "occurred_at", "data"}
	for _, key := range required {
		if _, exists := raw[key]; !exists {
			t.Fatalf("envelope missing required field %q in %s", key, payload)
		}
	}

	var got WSMessage
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("unmarshal canonical envelope: %v", err)
	}
	if got.ID != msg.ID {
		t.Fatalf("id = %q, want %q", got.ID, msg.ID)
	}
	if got.Version != CurrentVersion {
		t.Fatalf("version = %d, want %d", got.Version, CurrentVersion)
	}
	if got.Type != msg.Type {
		t.Fatalf("type = %q, want %q", got.Type, msg.Type)
	}
	if got.DrillID != msg.DrillID {
		t.Fatalf("drill_id = %d, want %d", got.DrillID, msg.DrillID)
	}
	if !got.OccurredAt.Equal(msg.OccurredAt) {
		t.Fatalf("occurred_at = %v, want %v", got.OccurredAt, msg.OccurredAt)
	}
	if string(got.Data) != string(msg.Data) {
		t.Fatalf("data = %s, want %s", got.Data, msg.Data)
	}
}

// TestSubscriberFansOutToClients verifies the subscriber decodes a WSMessage
// from Redis and invokes the handler — the single local fanout entry point.
func TestSubscriberFansOutToClients(t *testing.T) {
	occurred := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	original := WSMessage{
		ID:         "evt-fanout",
		Version:    CurrentVersion,
		Type:       "step.updated",
		DrillID:    42,
		OccurredAt: occurred,
		Data:       json.RawMessage(`{"step_id":7}`),
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	session := &fakePubSubSession{ch: make(chan *goredis.Message, 1)}
	session.ch <- &goredis.Message{Payload: string(encoded)}
	subscriber := &fakeRedisSubscriber{sessions: []*fakePubSubSession{session}}

	bus := &RedisBus{subscriber: subscriber, backoff: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var received atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- bus.Subscribe(ctx, func(msg WSMessage) {
			if msg.ID != original.ID || msg.Type != original.Type || msg.DrillID != original.DrillID {
				t.Errorf("handler got %+v, want %+v", msg, original)
			}
			if string(msg.Data) != string(original.Data) {
				t.Errorf("data = %s, want %s", msg.Data, original.Data)
			}
			received.Add(1)
		})
	}()

	waitFor(t, func() bool { return received.Load() > 0 && bus.Healthy() })

	cancel()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Subscribe did not return after context cancellation")
	}
}

// TestDuplicateDeliveryHarmless verifies that delivering the same event ID
// twice invokes the handler at most once — the subscriber dedups by ID.
func TestDuplicateDeliveryHarmless(t *testing.T) {
	original := WSMessage{
		ID:         "evt-dup",
		Version:    CurrentVersion,
		Type:       "step.updated",
		DrillID:    9,
		OccurredAt: time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
		Data:       json.RawMessage(`{"ok":true}`),
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	session := &fakePubSubSession{ch: make(chan *goredis.Message, 2)}
	session.ch <- &goredis.Message{Payload: string(encoded)}
	session.ch <- &goredis.Message{Payload: string(encoded)}

	subscriber := &fakeRedisSubscriber{sessions: []*fakePubSubSession{session}}
	bus := &RedisBus{subscriber: subscriber, backoff: time.Millisecond}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- bus.Subscribe(ctx, func(WSMessage) {
			calls.Add(1)
		})
	}()

	waitFor(t, func() bool { return bus.Healthy() })
	// Give the subscriber time to process both deliveries.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls.Load() >= 2 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("handler called %d times for duplicate event, want 1", got)
	}
}

// TestRedisSubscriptionLossMarksUnready verifies that when the Redis
// subscription channel closes (subscription loss), the subscriber marks the
// node unready (Healthy() returns false) and reconnects with backoff.
func TestRedisSubscriptionLossMarksUnready(t *testing.T) {
	first := &fakePubSubSession{ch: make(chan *goredis.Message)}
	second := &fakePubSubSession{ch: make(chan *goredis.Message)}
	subscriber := &fakeRedisSubscriber{sessions: []*fakePubSubSession{first, second}}
	// Use a 50ms backoff so the unready window is wide enough to observe —
	// with a 1ms backoff the subscriber reconnects before the poller sees it.
	bus := &RedisBus{subscriber: subscriber, backoff: 50 * time.Millisecond, seen: map[string]struct{}{}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- bus.Subscribe(ctx, func(WSMessage) {})
	}()

	// Wait for healthy on first session.
	waitFor(t, bus.Healthy)

	// Drop the subscription channel — simulates Redis subscription loss.
	close(first.ch)
	waitFor(t, first.isClosed)

	// Healthy must report false after the subscription is lost (even briefly
	// before reconnect succeeds).
	deadline := time.Now().Add(300 * time.Millisecond)
	sawUnready := false
	for time.Now().Before(deadline) {
		if !bus.Healthy() {
			sawUnready = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !sawUnready {
		t.Fatal("subscriber remained healthy after subscription loss")
	}

	// Reconnect restores healthy state.
	waitFor(t, func() bool { return subscriber.callCount() >= 2 && bus.Healthy() })

	cancel()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Subscribe did not return after context cancellation")
	}
}
