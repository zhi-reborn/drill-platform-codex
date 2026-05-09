package flowengine

import (
	"fmt"
	"sync"
	"time"
)

type timeoutEntry struct {
	flowInstID  int64
	stepDefID   int64
	stepInstID  int64
	timeoutAt   time.Time
}

type TimeoutScheduler struct {
	mu      sync.Mutex
	entries map[string]*timeoutEntry
	eventBus *EventBus
	stopCh  chan struct{}
	running bool
}

func NewTimeoutScheduler(eventBus *EventBus) *TimeoutScheduler {
	return &TimeoutScheduler{
		entries:  make(map[string]*timeoutEntry),
		eventBus: eventBus,
		stopCh:   make(chan struct{}),
	}
}

func (ts *TimeoutScheduler) Start() {
	ts.mu.Lock()
	if ts.running {
		ts.mu.Unlock()
		return
	}
	ts.running = true
	ts.mu.Unlock()
	go ts.scanLoop()
}

func (ts *TimeoutScheduler) Stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.running {
		return
	}
	ts.running = false
	close(ts.stopCh)
	ts.stopCh = make(chan struct{})
}

func (ts *TimeoutScheduler) Register(flowInstID, stepDefID, stepInstID int64, timeoutAt time.Time) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	key := ts.key(flowInstID, stepDefID)
	ts.entries[key] = &timeoutEntry{
		flowInstID: flowInstID,
		stepDefID:  stepDefID,
		stepInstID: stepInstID,
		timeoutAt:  timeoutAt,
	}
}

func (ts *TimeoutScheduler) Unregister(flowInstID, stepDefID int64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	key := ts.key(flowInstID, stepDefID)
	delete(ts.entries, key)
}

func (ts *TimeoutScheduler) scanLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ts.stopCh:
			return
		case <-ticker.C:
			ts.scan()
		}
	}
}

func (ts *TimeoutScheduler) scan() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now()
	var expired []*timeoutEntry
	for key, entry := range ts.entries {
		if now.After(entry.timeoutAt) {
			expired = append(expired, entry)
			delete(ts.entries, key)
		}
	}
	for _, entry := range expired {
		ts.eventBus.emit(EventStepTimeout, entry.flowInstID, entry.stepInstID, entry.stepDefID, map[string]interface{}{
			"timeout_at": entry.timeoutAt,
		})
	}
}

func (ts *TimeoutScheduler) key(flowInstID, stepDefID int64) string {
	return fmt.Sprintf("%d_%d", flowInstID, stepDefID)
}
