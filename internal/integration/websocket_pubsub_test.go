//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"drill-platform/internal/infrastructure/events"
	"drill-platform/internal/infrastructure/websocket"
)

// TestCrossNodeWebSocketEventDelivery verifies that an event published through
// Redis Pub/Sub on one connection (backend A) is delivered to a WebSocket
// client connected to a different manager (backend B) via a separate Redis
// subscription. This proves the cross-node fan-out path without requiring two
// HTTP servers.
func TestCrossNodeWebSocketEventDelivery(t *testing.T) {
	rdbA := connectRedisRaw(t)
	rdbB := connectRedisRaw(t)
	defer rdbA.Close()
	defer rdbB.Close()

	busA := events.NewRedisBus(rdbA)
	busB := events.NewRedisBus(rdbB)

	// Manager B represents backend B; it will receive events via Redis and
	// deliver them to locally connected WebSocket clients.
	managerB := websocket.NewManager()
	go managerB.Run()

	// Subscribe manager B's DeliverEvent handler to the Redis bus.
	subCtx, subCancel := context.WithCancel(context.Background())
	defer subCancel()
	go func() {
		_ = busB.Subscribe(subCtx, managerB.DeliverEvent)
	}()

	// Wait for the subscriber to become healthy so we don't publish before the
	// subscription is active.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if busB.Healthy() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !busB.Healthy() {
		t.Fatal("bus B subscriber did not become healthy")
	}

	// Register a WebSocket client on manager B subscribed to drill 42.
	// The client uses a nil connection; only the Send channel is read in the test.
	const drillID = uint(42)
	clientB := websocket.NewClient("client-b", 0, nil, drillID, websocket.ChannelDisplay)
	managerB.Mu.Lock()
	managerB.Channels[websocket.ChannelDisplay].AddClient(clientB)
	managerB.Mu.Unlock()

	// Publish a drill event through bus A (simulating backend A executing a
	// command and publishing the resulting WebSocket event).
	payload := json.RawMessage(`{"type":"step.updated","status":"completed"}`)
	event := events.NewWSMessage("step.updated", uint64(drillID), 0, payload)

	pubCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pubCancel()
	if err := busA.Publish(pubCtx, event); err != nil {
		t.Fatalf("publish event through bus A: %v", err)
	}

	// Assert the client on manager B receives the canonical WSMessage.
	select {
	case got := <-clientB.Send:
		if got.ID != event.ID || got.Type != event.Type || got.DrillID != event.DrillID {
			t.Fatalf("received event = %+v, want %+v", got, event)
		}
		if string(got.Data) != string(payload) {
			t.Fatalf("received data = %s, want %s", got.Data, payload)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("client B did not receive the cross-node event")
	}

	// Verify a non-matching drill does not receive the event. Publish an event
	// for drill 99 and confirm client B (drill 42) does not get it.
	otherEvent := events.NewWSMessage("step.updated", 99, 0, json.RawMessage(`{"type":"step.updated","status":"skipped"}`))
	if err := busA.Publish(pubCtx, otherEvent); err != nil {
		t.Fatalf("publish other-drill event: %v", err)
	}
	select {
	case got := <-clientB.Send:
		t.Fatalf("client B received an event for a non-matching drill: %+v", got)
	case <-time.After(500 * time.Millisecond):
		// Expected: no event for drill 99 arrives at the drill 42 client.
	}
}
