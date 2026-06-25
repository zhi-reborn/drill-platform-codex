package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/infrastructure/events"
)

func addClientForEventTest(m *Manager, c *Client) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Channels[c.Type].AddClient(c)
}

func assertReceivedEventPayload(t *testing.T, c *Client, want []byte) {
	t.Helper()
	select {
	case got := <-c.Send:
		if string(got) != string(want) {
			t.Fatalf("received payload = %s, want %s", got, want)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client did not receive event payload")
	}
}

func assertNoEventPayload(t *testing.T, c *Client) {
	t.Helper()
	select {
	case got := <-c.Send:
		t.Fatalf("unexpected payload received: %s", got)
	case <-time.After(20 * time.Millisecond):
	}
}

type fakeEventBus struct {
	mu       sync.Mutex
	handlers []func(events.Event)
}

func (b *fakeEventBus) Publish(ctx context.Context, event events.Event) error {
	b.mu.Lock()
	handlers := append([]func(events.Event){}, b.handlers...)
	b.mu.Unlock()
	for _, handler := range handlers {
		handler(event)
	}
	return nil
}

func (b *fakeEventBus) Subscribe(ctx context.Context, handler func(events.Event)) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
	return nil
}

func (b *fakeEventBus) Healthy() bool { return true }

func TestEventBusBroadcastsDrillEventToMatchingClientsAcrossManagers(t *testing.T) {
	payload := json.RawMessage(`{"type":"step.updated"}`)
	event := events.Event{ID: "evt-1", Type: "step.updated", DrillID: 42, Payload: payload}

	managerA := NewManager()
	managerB := NewManager()
	matchingA := NewClient("display-a", 0, nil, 42, ChannelDisplay)
	matchingB := NewClient("control-b", 0, nil, 42, ChannelControl)
	nonMatching := NewClient("display-other", 0, nil, 43, ChannelDisplay)
	addClientForEventTest(managerA, matchingA)
	addClientForEventTest(managerA, nonMatching)
	addClientForEventTest(managerB, matchingB)

	bus := &fakeEventBus{}
	if err := bus.Subscribe(context.Background(), managerA.DeliverEvent); err != nil {
		t.Fatalf("subscribe manager A: %v", err)
	}
	if err := bus.Subscribe(context.Background(), managerB.DeliverEvent); err != nil {
		t.Fatalf("subscribe manager B: %v", err)
	}
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	assertReceivedEventPayload(t, matchingA, payload)
	assertReceivedEventPayload(t, matchingB, payload)
	assertNoEventPayload(t, nonMatching)
}

func TestDeliverEventBroadcastsUserEventOnlyToMatchingTaskClients(t *testing.T) {
	payload := json.RawMessage(`{"type":"task.updated"}`)
	event := events.Event{ID: "evt-2", Type: "task.updated", UserID: 9, Payload: payload}
	manager := NewManager()
	matching := NewClient("tasks-9", 9, nil, 0, ChannelTasks)
	nonMatching := NewClient("tasks-10", 10, nil, 0, ChannelTasks)
	displaySameUser := NewClient("display-9", 9, nil, 0, ChannelDisplay)
	addClientForEventTest(manager, matching)
	addClientForEventTest(manager, nonMatching)
	addClientForEventTest(manager, displaySameUser)

	manager.DeliverEvent(event)

	assertReceivedEventPayload(t, matching, payload)
	assertNoEventPayload(t, nonMatching)
	assertNoEventPayload(t, displaySameUser)
}
