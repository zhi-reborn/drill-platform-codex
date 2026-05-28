package websocket

import "encoding/json"

func (m *Manager) BroadcastMessage(msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	m.Broadcast <- data
	return nil
}

func (m *Manager) BroadcastToDrill(drillID uint, msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	m.BroadcastToDrillRaw(drillID, data)
	return nil
}

func (m *Manager) BroadcastToDrillRaw(drillID uint, data []byte) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	for _, ch := range m.Channels {
		clients := ch.GetClientsByDrill(drillID)
		for _, c := range clients {
			select {
			case c.Send <- data:
			default:
				close(c.Send)
				delete(ch.clients, c.ID)
			}
		}
	}
}

func (m *Manager) BroadcastToUser(userID uint, msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	m.BroadcastToUserRaw(userID, data)
	return nil
}

func (m *Manager) BroadcastToUserRaw(userID uint, data []byte) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	if ch, ok := m.Channels[ChannelTasks]; ok {
		clients := ch.GetClientsByUser(userID)
		for _, c := range clients {
			select {
			case c.Send <- data:
			default:
				close(c.Send)
				ch.RemoveClient(c)
			}
		}
	}
}

func (m *Manager) SendStepChange(drillID uint, payload StepChangePayload) {
	eventType := stepStatusToEvent(payload.NewStatus)
	msg := NewMessage(eventType, drillID, payload)
	m.BroadcastToDrill(drillID, msg)
}

func (m *Manager) SendTimeoutWarning(drillID uint, userID uint, payload TimeoutWarningPayload) {
	msg := NewMessage(EventTimeoutWarning, drillID, payload)
	m.BroadcastToDrill(drillID, msg)
	m.BroadcastToUser(userID, msg)
}

func (m *Manager) SendDrillStatus(drillID uint, payload DrillStatusPayload) {
	eventType := drillStatusToEvent(payload.NewStatus, payload.PreviousStatus)
	msg := NewMessage(eventType, drillID, payload)
	m.BroadcastToDrill(drillID, msg)
}

func (m *Manager) SendControlEvent(drillID uint, payload ControlPayload) {
	eventType := controlActionToEvent(payload.Action)
	msg := NewMessage(eventType, drillID, payload)
	m.BroadcastToDrill(drillID, msg)
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

func (m *Manager) SendTaskUpdate(userID uint, payload TaskAssignPayload) {
	msg := NewMessage(EventInfo, 0, payload)
	m.BroadcastToUser(userID, msg)
}
