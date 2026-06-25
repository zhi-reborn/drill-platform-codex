package websocket

import (
	"encoding/json"
	"log"
	"time"

	"drill-platform/internal/infrastructure/events"
)

func (m *Manager) StartHeartbeat() {
	ticker := time.NewTicker(PingPeriod)
	go func() {
		for range ticker.C {
			m.checkClientHeartbeats()
		}
	}()
}

// checkClientHeartbeats evicts clients whose lastPong is older than PongWait.
// Stale clients are routed through the Unregister channel so the hub goroutine
// serializes Send-channel closure with any in-flight broadcast — closing Send
// here directly would race with processBroadcast writing to the same channel.
func (m *Manager) checkClientHeartbeats() {
	m.Mu.Lock()
	now := time.Now()
	var stale []*Client
	for _, ch := range m.Channels {
		for id, c := range ch.clients {
			if now.Sub(c.lastPong) > PongWait {
				log.Printf("客户端心跳超时 [%s], 关闭连接", id)
				ch.RemoveClient(c)
				stale = append(stale, c)
			}
		}
	}
	m.Mu.Unlock()

	for _, c := range stale {
		// Mark closed so SendJSON fast-fails; the hub will closeSend when it
		// processes the Unregister job. Close is idempotent.
		c.Close()
		select {
		case m.Unregister <- c:
		default:
			// Unregister buffer full — fall back to closeSend so WritePump
			// still terminates. The hub may process Unregister later and
			// call closeSend again (idempotent via sendOnce).
			c.closeSend()
		}
	}
}

func (m *Manager) handleClientMessage(c *Client, message []byte) {
	var msg events.WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		_ = c.SendJSON(NewErrorMessage("消息格式错误"))
		return
	}

	switch msg.Type {
	case EventPing:
		_ = c.SendJSON(NewPongMessage())
	default:
	}
}
