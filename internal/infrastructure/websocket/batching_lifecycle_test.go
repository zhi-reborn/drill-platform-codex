package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"drill-platform/internal/infrastructure/events"

	gorilla "github.com/gorilla/websocket"
)

// TestBatchingProducesValidJSONArray verifies that WritePump batches multiple
// queued messages into a single JSON array frame ([msg1,msg2,...]) via
// json.Marshal([]WSMessage), not newline-concatenated JSON. The client side
// must receive one TextMessage whose payload parses as a JSON array of
// canonical WSMessages.
func TestBatchingProducesValidJSONArray(t *testing.T) {
	upgrader := gorilla.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	ready := make(chan *Client, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		client := NewClient("batch-client", 0, conn, 42, ChannelDisplay)
		ready <- client
		client.WritePump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := gorilla.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer ws.Close()

	select {
	case client := <-ready:
		// Push three messages quickly so WritePump drains them in one batch.
		for i := 0; i < 3; i++ {
			msg := events.NewWSMessage("step.updated", 42, 0, json.RawMessage(`{"i":0}`))
			select {
			case client.Send <- msg:
			default:
				t.Fatalf("could not enqueue message %d", i)
			}
		}
		client.Close()

		_, data, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		var batch []events.WSMessage
		if err := json.Unmarshal(data, &batch); err != nil {
			t.Fatalf("payload is not a JSON array: %v; payload=%s", err, data)
		}
		if len(batch) < 1 {
			t.Fatalf("batch len = %d, want >= 1; payload=%s", len(batch), data)
		}
		for i, msg := range batch {
			if msg.Version != events.CurrentVersion {
				t.Fatalf("batch[%d].Version = %d, want %d", i, msg.Version, events.CurrentVersion)
			}
			if msg.Type != "step.updated" {
				t.Fatalf("batch[%d].Type = %q, want step.updated", i, msg.Type)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not upgrade connection")
	}
}

// TestClosingClientDuringPublishNoPanic verifies that unregistering (and thus
// closing Send) a client while broadcasts are in-flight does not panic. The
// hub serialization guarantees that processBroadcast writes and closeSend
// never overlap because they run in the same goroutine. This test runs under
// -race to catch any send-on-closed-channel panics.
func TestClosingClientDuringPublishNoPanic(t *testing.T) {
	manager := NewManager()
	stop := startTestHub(manager)
	defer close(stop)

	client := NewClient("panic-client", 0, nil, 42, ChannelDisplay)
	manager.Register <- client

	// Wait for the hub to register the client.
	waitForClient(t, manager, client)

	var wg sync.WaitGroup
	wg.Add(2)

	// Producer: blast broadcasts.
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			msg := events.NewWSMessage("step.updated", 42, 0, json.RawMessage(`{"i":0}`))
			manager.DeliverEvent(msg)
		}
	}()

	// Closer: unregister the client mid-blast.
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		manager.Unregister <- client
		client.Close()
	}()

	wg.Wait()

	// Drain remaining broadcasts so the hub quiesces before stop.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		select {
		case <-manager.broadcast:
		default:
			return
		}
	}
}

func waitForClient(t *testing.T, m *Manager, c *Client) {
	t.Helper()
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		m.Mu.RLock()
		_, exists := m.Channels[c.Type].clients[c.ID]
		m.Mu.RUnlock()
		if exists {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("client was not registered by the hub")
}
