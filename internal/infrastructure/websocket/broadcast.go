package websocket

import "encoding/json"

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
	// 先收集需要慢速删除的客户端，避免在 RLock 下修改 map
	var slowClients []*Client

	for _, ch := range m.Channels {
		clients := ch.GetClientsByDrill(drillID)
		for _, c := range clients {
			select {
			case c.Send <- data:
			default:
				// 慢客户端：不关闭 Send channel（避免 WritePump panic），只标记待移除
				slowClients = append(slowClients, c)
			}
		}
	}
	m.Mu.RUnlock()

	// 释放 RLock 后，再处理慢客户端移除（需要写锁）
	if len(slowClients) > 0 {
		m.Mu.Lock()
		for _, c := range slowClients {
			if ch, ok := m.Channels[c.Type]; ok {
				ch.RemoveClient(c)
			}
		}
		m.Mu.Unlock()
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
	var slowClients []*Client

	if ch, ok := m.Channels[ChannelTasks]; ok {
		clients := ch.GetClientsByUser(userID)
		for _, c := range clients {
			select {
			case c.Send <- data:
			default:
				slowClients = append(slowClients, c)
			}
		}
	}
	m.Mu.RUnlock()

	if len(slowClients) > 0 {
		m.Mu.Lock()
		for _, c := range slowClients {
			if ch, ok := m.Channels[c.Type]; ok {
				ch.RemoveClient(c)
			}
		}
		m.Mu.Unlock()
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
