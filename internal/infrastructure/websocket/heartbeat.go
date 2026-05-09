package websocket

import (
	"encoding/json"
	"log"
	"time"
)

func (m *Manager) StartHeartbeat() {
	ticker := time.NewTicker(PingPeriod)
	go func() {
		for range ticker.C {
			m.checkClientHeartbeats()
		}
	}()
}

func (m *Manager) checkClientHeartbeats() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	now := time.Now()

	for _, ch := range m.Channels {
		for id, c := range ch.clients {
			if now.Sub(c.lastPong) > PongWait {
				log.Printf("客户端心跳超时 [%s], 关闭连接", id)
				c.Close()
				ch.RemoveClient(c)
			}
		}
	}
}

func (m *Manager) handleClientMessage(c *Client, message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.SendJSON(NewErrorMessage("消息格式错误"))
		return
	}

	switch msg.EventType {
	case EventPing:
		c.SendJSON(NewPongMessage())
	default:
	}
}
