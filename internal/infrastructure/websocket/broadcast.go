package websocket

import (
	"encoding/json"
	"log"

	"drill-platform/internal/infrastructure/events"
)

// DeliverEvent is the single local fanout entry point. The Redis subscriber
// calls this with each deduped WSMessage; it enqueues a broadcast job onto
// the hub channel so all client writes are serialized through the hub
// goroutine (no concurrent Send writes/closures).
func (m *Manager) DeliverEvent(msg events.WSMessage) {
	if msg.DrillID != 0 {
		m.enqueueBroadcast(broadcastJob{drillID: uint(msg.DrillID), msg: msg})
	}
	if msg.UserID != 0 {
		m.enqueueBroadcast(broadcastJob{userID: uint(msg.UserID), msg: msg})
	}
}

// enqueueBroadcast sends a broadcast job to the hub goroutine. If the
// broadcast channel is full (clients backpressured) the event is dropped with
// a log line — the next event will retry.
func (m *Manager) enqueueBroadcast(job broadcastJob) {
	select {
	case m.broadcast <- job:
	default:
		log.Printf("broadcast channel full, dropping event type=%s id=%s", job.msg.Type, job.msg.ID)
	}
}

// BroadcastToDrill enqueues a broadcast of msg to all clients subscribed to
// drillID. Used by the legacy Send* helpers when no Redis publisher is wired
// (dev/test fallback). In production, events go through Redis and arrive via
// DeliverEvent.
func (m *Manager) BroadcastToDrill(drillID uint, msg events.WSMessage) error {
	m.enqueueBroadcast(broadcastJob{drillID: drillID, msg: msg})
	return nil
}

// BroadcastToUser enqueues a broadcast of msg to all task-channel clients for
// userID. See BroadcastToDrill for the dev/test fallback note.
func (m *Manager) BroadcastToUser(userID uint, msg events.WSMessage) error {
	m.enqueueBroadcast(broadcastJob{userID: userID, msg: msg})
	return nil
}

func (m *Manager) SendStepChange(drillID uint, payload StepChangePayload) {
	eventType := stepStatusToEvent(payload.NewStatus)
	m.publishOrDeliver(eventType, drillID, 0, payload)
}

func (m *Manager) SendTimeoutWarning(drillID uint, userID uint, payload TimeoutWarningPayload) {
	m.publishOrDeliver(EventTimeoutWarning, drillID, userID, payload)
}

func (m *Manager) SendDrillStatus(drillID uint, payload DrillStatusPayload) {
	eventType := drillStatusToEvent(payload.NewStatus, payload.PreviousStatus)
	m.publishOrDeliver(eventType, drillID, 0, payload)
}

func (m *Manager) SendControlEvent(drillID uint, payload ControlPayload) {
	eventType := controlActionToEvent(payload.Action)
	m.publishOrDeliver(eventType, drillID, 0, payload)
}

func (m *Manager) SendTaskUpdate(userID uint, payload TaskAssignPayload) {
	m.publishOrDeliver(EventInfo, 0, userID, payload)
}

// publishOrDeliver is the single publication path: if a Redis publisher is
// wired, the event goes to Redis (the subscriber will call DeliverEvent to
// fan out locally). With no publisher (dev/test without Redis), it falls
// back to local DeliverEvent so single-node mode still works.
func (m *Manager) publishOrDeliver(eventType string, drillID, userID uint, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal ws payload type=%s: %v", eventType, err)
		return
	}
	msg := events.NewWSMessage(eventType, uint64(drillID), uint64(userID), data)
	if m.publisher != nil {
		if err := m.publisher.Publish(m.publishCtx, msg); err != nil {
			log.Printf("publish event type=%s: %v", eventType, err)
		}
		return
	}
	m.DeliverEvent(msg)
}

func drillStatusToEvent(newStatus, prevStatus string) string {
	switch newStatus {
	case "running":
		if prevStatus == "paused" {
			return EventDrillResumed
		}
		return EventDrillStarted
	case "paused":
		return EventDrillPaused
	case "completed":
		return EventDrillCompleted
	case "terminated":
		return EventDrillTerminated
	default:
		return EventDrillStarted
	}
}

func stepStatusToEvent(status string) string {
	switch status {
	case "running":
		return EventStepStarted
	case "completed":
		return EventStepComplete
	case "timeout":
		return EventStepTimeout
	case "skipped":
		return EventStepSkipped
	case "issue":
		return EventStepIssue
	default:
		return EventStepComplete
	}
}

func controlActionToEvent(action string) string {
	switch action {
	case "pause":
		return EventControlPause
	case "resume":
		return EventControlResume
	case "terminate":
		return EventControlTerminate
	case "comment":
		return EventControlComment
	default:
		return EventControlPause
	}
}
