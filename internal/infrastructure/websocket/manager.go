package websocket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Manager struct {
	Channels   map[string]*Channel
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		Channels: map[string]*Channel{
			ChannelDisplay: NewChannel(ChannelDisplay),
			ChannelTasks:   NewChannel(ChannelTasks),
			ChannelControl: NewChannel(ChannelControl),
		},
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
	}
}

func (m *Manager) Run() {
	m.StartHeartbeat()

	for {
		select {
		case client := <-m.Register:
			m.Mu.Lock()
			if ch, ok := m.Channels[client.Type]; ok {
				ch.AddClient(client)
				log.Printf("客户端 [%s] 注册到通道 [%s], drillID: %d, userID: %d",
					client.ID, client.Type, client.DrillID, client.UserID)
			}
			m.Mu.Unlock()

		case client := <-m.Unregister:
			m.Mu.Lock()
			if ch, ok := m.Channels[client.Type]; ok {
				ch.RemoveClient(client)
				log.Printf("客户端 [%s] 从通道 [%s] 注销", client.ID, client.Type)
			}
			m.Mu.Unlock()

		case message := <-m.Broadcast:
			m.Mu.RLock()
			for _, ch := range m.Channels {
				for _, client := range ch.clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						ch.RemoveClient(client)
					}
				}
			}
			m.Mu.RUnlock()
		}
	}
}

func (m *Manager) GetStats() map[string]interface{} {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	stats := make(map[string]interface{})
	for name, ch := range m.Channels {
		stats[name] = ch.ClientCount()
	}
	return stats
}

func extractUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	switch v := userID.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	default:
		return 0
	}
}

func generateClientID() string {
	return fmt.Sprintf("client-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
