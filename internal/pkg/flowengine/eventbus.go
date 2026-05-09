package flowengine

import (
	"sync"
	"time"
)

// EventBus 事件总线，用于引擎内部与外部组件（WebSocket Handler 等）解耦
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
}

func (eb *EventBus) SubscribeAll(ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for et := range eb.subscribers {
		eb.subscribers[et] = append(eb.subscribers[et], ch)
	}
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if chs, ok := eb.subscribers[event.Type]; ok {
		for _, ch := range chs {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (eb *EventBus) Unsubscribe(eventType EventType, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	chs := eb.subscribers[eventType]
	for i, c := range chs {
		if c == ch {
			eb.subscribers[eventType] = append(chs[:i], chs[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) emit(t EventType, flowInstID, stepInstID, stepDefID int64, payload interface{}) {
	eb.Publish(Event{
		Type:       t,
		FlowInstID: flowInstID,
		StepInstID: stepInstID,
		StepDefID:  stepDefID,
		Payload:    payload,
		Timestamp:  time.Now(),
	})
}
