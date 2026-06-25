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

// startTestHub launches a minimal hub loop that drains the broadcast channel
// and processes Register/Unregister jobs. It returns a stop channel that
// cancels the loop when closed. This mirrors production semantics where the
// hub goroutine serializes all client writes — without it, DeliverEvent would
// enqueue jobs that nobody consumes.
func startTestHub(m *Manager) chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case client := <-m.Register:
				m.Mu.Lock()
				if ch, ok := m.Channels[client.Type]; ok {
					ch.AddClient(client)
				}
				m.Mu.Unlock()
			case client := <-m.Unregister:
				m.Mu.Lock()
				if ch, ok := m.Channels[client.Type]; ok {
					ch.RemoveClient(client)
				}
				m.Mu.Unlock()
				client.closeSend()
			case job := <-m.broadcast:
				m.processBroadcast(job)
			}
		}
	}()
	return done
}

func assertReceivedEvent(t *testing.T, c *Client, want events.WSMessage) {
	t.Helper()
	select {
	case got := <-c.Send:
		if got.ID != want.ID || got.Type != want.Type || got.DrillID != want.DrillID {
			t.Fatalf("received event = %+v, want %+v", got, want)
		}
		if string(got.Data) != string(want.Data) {
			t.Fatalf("received data = %s, want %s", got.Data, want.Data)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client did not receive event")
	}
}

func assertNoEvent(t *testing.T, c *Client) {
	t.Helper()
	select {
	case got := <-c.Send:
		t.Fatalf("unexpected event received: %+v", got)
	case <-time.After(30 * time.Millisecond):
	}
}

// fakeEventBus is a minimal events.Publisher + events.Subscriber stub that
// synchronously dispatches published WSMessages to all registered handlers.
type fakeEventBus struct {
	mu       sync.Mutex
	handlers []func(events.WSMessage)
}

func (b *fakeEventBus) Publish(_ context.Context, msg events.WSMessage) error {
	b.mu.Lock()
	handlers := append([]func(events.WSMessage){}, b.handlers...)
	b.mu.Unlock()
	for _, handler := range handlers {
		handler(msg)
	}
	return nil
}

func (b *fakeEventBus) Subscribe(_ context.Context, handler func(events.WSMessage)) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
	return nil
}

func (b *fakeEventBus) Healthy() bool { return true }

func TestEventBusBroadcastsDrillEventToMatchingClientsAcrossManagers(t *testing.T) {
	managerA := NewManager()
	managerB := NewManager()
	stopA := startTestHub(managerA)
	stopB := startTestHub(managerB)
	defer close(stopA)
	defer close(stopB)

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

	msg := events.NewWSMessage("step.updated", 42, 0, json.RawMessage(`{"type":"step.updated"}`))
	if err := bus.Publish(context.Background(), msg); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	assertReceivedEvent(t, matchingA, msg)
	assertReceivedEvent(t, matchingB, msg)
	assertNoEvent(t, nonMatching)
}

func TestDeliverEventBroadcastsUserEventOnlyToMatchingTaskClients(t *testing.T) {
	manager := NewManager()
	stop := startTestHub(manager)
	defer close(stop)

	matching := NewClient("tasks-9", 9, nil, 0, ChannelTasks)
	nonMatching := NewClient("tasks-10", 10, nil, 0, ChannelTasks)
	displaySameUser := NewClient("display-9", 9, nil, 0, ChannelDisplay)
	addClientForEventTest(manager, matching)
	addClientForEventTest(manager, nonMatching)
	addClientForEventTest(manager, displaySameUser)

	msg := events.NewWSMessage("task.updated", 0, 9, json.RawMessage(`{"type":"task.updated"}`))
	manager.DeliverEvent(msg)

	assertReceivedEvent(t, matching, msg)
	assertNoEvent(t, nonMatching)
	assertNoEvent(t, displaySameUser)
}
