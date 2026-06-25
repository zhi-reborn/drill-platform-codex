package websocket

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"drill-platform/internal/infrastructure/events"

	"github.com/gin-gonic/gin"
)

// broadcastJob is a serialized unit of fanout work processed by the hub
// goroutine in Run(). drillID and userID are mutually exclusive targeting
// keys — drillID fans out to display/control channels, userID to tasks.
type broadcastJob struct {
	drillID uint
	userID  uint
	msg     events.WSMessage
}

// Manager owns the client lifecycle. Register, Unregister, and broadcast
// jobs are all funneled through channels and processed serially by the Run
// goroutine. This serialization is the key safety property: only the hub
// goroutine writes to (and closes) Client.Send channels, so concurrent
// publish + close never panics.
type Manager struct {
	Channels   map[string]*Channel
	Register   chan *Client
	Unregister chan *Client
	broadcast  chan broadcastJob
	Mu         sync.RWMutex

	// publisher, when set, routes Send* events through Redis. The subscriber
	// half calls DeliverEvent to fan out locally. Nil = dev/test fallback
	// (DeliverEvent broadcasts directly to the local hub).
	publisher  events.Publisher
	publishCtx context.Context
}

func NewManager() *Manager {
	return &Manager{
		Channels: map[string]*Channel{
			ChannelDisplay: NewChannel(ChannelDisplay),
			ChannelTasks:   NewChannel(ChannelTasks),
			ChannelControl: NewChannel(ChannelControl),
		},
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
		broadcast:  make(chan broadcastJob, 512),
		publishCtx: context.Background(),
	}
}

// SetPublisher wires a Redis publisher. When set, the Send* helpers publish
// to Redis instead of broadcasting locally; the subscriber (started in
// main.go) calls DeliverEvent to fan out. When nil (no Redis), Send* falls
// back to local DeliverEvent so single-node dev/test still works.
func (m *Manager) SetPublisher(p events.Publisher) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.publisher = p
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
			// Hub owns Send closure — closing here (serialized with
			// broadcasts) guarantees no concurrent write+close panic.
			client.closeSend()

		case job := <-m.broadcast:
			m.processBroadcast(job)
		}
	}
}

// processBroadcast fans a single message out to matching clients. It runs in
// the hub goroutine, so Send writes here cannot race with Send closure in
// the Unregister case. Slow clients (full Send buffer) are queued for
// removal after releasing the read lock.
func (m *Manager) processBroadcast(job broadcastJob) {
	m.Mu.RLock()
	var slow []*Client
	for _, ch := range m.Channels {
		var clients []*Client
		if job.drillID != 0 {
			clients = ch.GetClientsByDrill(job.drillID)
		} else if job.userID != 0 {
			clients = ch.GetClientsByUser(job.userID)
		}
		for _, c := range clients {
			select {
			case c.Send <- job.msg:
			default:
				slow = append(slow, c)
			}
		}
	}
	m.Mu.RUnlock()

	if len(slow) > 0 {
		m.Mu.Lock()
		for _, c := range slow {
			if ch, ok := m.Channels[c.Type]; ok {
				ch.RemoveClient(c)
			}
		}
		m.Mu.Unlock()
		for _, c := range slow {
			c.closeSend()
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
	// JWTAuth 中间件将 userID 存入 "user_id_int"（uint64 类型）
	if v, exists := c.Get("user_id_int"); exists {
		switch val := v.(type) {
		case uint64:
			return uint(val)
		case uint:
			return val
		case float64:
			return uint(val)
		}
	}
	return 0
}

func generateClientID() string {
	return fmt.Sprintf("client-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
