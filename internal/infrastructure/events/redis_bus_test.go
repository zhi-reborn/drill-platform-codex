package events

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type fakeRedisClient struct {
	channel string
	message interface{}
}

func (f *fakeRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	f.channel = channel
	f.message = message
	return nil
}

func TestRedisBusPublishJSONUsesEventsChannel(t *testing.T) {
	client := &fakeRedisClient{}
	bus := NewRedisBus(client)
	event := Event{
		ID:        "evt-1",
		Type:      "step.updated",
		DrillID:   7,
		Payload:   json.RawMessage(`{"ok":true}`),
		CreatedAt: time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC),
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if client.channel != EventsChannel {
		t.Fatalf("Publish() channel = %q, want %q", client.channel, EventsChannel)
	}

	payload, ok := client.message.(string)
	if !ok {
		t.Fatalf("Publish() message type = %T, want string", client.message)
	}

	var got Event
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("Publish() message is not valid event JSON: %v", err)
	}
	if got.ID != event.ID || got.Type != event.Type || got.DrillID != event.DrillID || string(got.Payload) != string(event.Payload) || !got.CreatedAt.Equal(event.CreatedAt) {
		t.Fatalf("Publish() event = %+v, want %+v", got, event)
	}
}

func TestRedisBusHandleMessageSkipsInvalidJSON(t *testing.T) {
	bus := NewRedisBus(&fakeRedisClient{})
	called := false

	bus.handleMessage("not-json", func(Event) {
		called = true
	})

	if called {
		t.Fatal("handler was called for invalid JSON")
	}
}

type fakeRedisSubscriber struct {
	mu       sync.Mutex
	sessions []*fakePubSubSession
	calls    int
}

func (f *fakeRedisSubscriber) Subscribe(ctx context.Context, channels ...string) pubSubSession {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if len(channels) != 1 || channels[0] != EventsChannel {
		return &fakePubSubSession{receiveErr: errors.New("unexpected channel"), ch: make(chan *goredis.Message)}
	}
	if len(f.sessions) == 0 {
		return &fakePubSubSession{receiveErr: errors.New("no session"), ch: make(chan *goredis.Message)}
	}
	s := f.sessions[0]
	f.sessions = f.sessions[1:]
	return s
}

func (f *fakeRedisSubscriber) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type fakePubSubSession struct {
	receiveErr error
	ch         chan *goredis.Message
	closed     bool
	mu         sync.Mutex
}

func (s *fakePubSubSession) Receive(ctx context.Context) (interface{}, error) {
	return nil, s.receiveErr
}

func (s *fakePubSubSession) Channel() <-chan *goredis.Message {
	return s.ch
}

func (s *fakePubSubSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *fakePubSubSession) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func TestRedisBusSubscribeReconnectsAndUpdatesHealthy(t *testing.T) {
	first := &fakePubSubSession{ch: make(chan *goredis.Message)}
	second := &fakePubSubSession{ch: make(chan *goredis.Message)}
	subscriber := &fakeRedisSubscriber{sessions: []*fakePubSubSession{first, second}}
	bus := &RedisBus{subscriber: subscriber, backoff: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- bus.Subscribe(ctx, func(Event) {})
	}()

	waitFor(t, func() bool { return bus.Healthy() })
	close(first.ch)
	waitFor(t, first.isClosed)
	waitFor(t, func() bool { return subscriber.callCount() >= 2 && bus.Healthy() })

	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Subscribe did not return after context cancellation")
	}
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met")
}
