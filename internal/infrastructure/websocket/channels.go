package websocket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Channel struct {
	Type      string
	clients   map[string]*Client
	subrooms  map[uint]map[string]*Client
	userTasks map[uint]map[string]*Client
}

func NewChannel(channelType string) *Channel {
	return &Channel{
		Type:      channelType,
		clients:   make(map[string]*Client),
		subrooms:  make(map[uint]map[string]*Client),
		userTasks: make(map[uint]map[string]*Client),
	}
}

func (ch *Channel) AddClient(c *Client) {
	ch.clients[c.ID] = c

	switch ch.Type {
	case ChannelDisplay, ChannelControl:
		if ch.subrooms[c.DrillID] == nil {
			ch.subrooms[c.DrillID] = make(map[string]*Client)
		}
		ch.subrooms[c.DrillID][c.ID] = c

	case ChannelTasks:
		if ch.userTasks[c.UserID] == nil {
			ch.userTasks[c.UserID] = make(map[string]*Client)
		}
		ch.userTasks[c.UserID][c.ID] = c
	}
}

func (ch *Channel) RemoveClient(c *Client) {
	delete(ch.clients, c.ID)

	switch ch.Type {
	case ChannelDisplay, ChannelControl:
		if room, ok := ch.subrooms[c.DrillID]; ok {
			delete(room, c.ID)
			if len(room) == 0 {
				delete(ch.subrooms, c.DrillID)
			}
		}

	case ChannelTasks:
		if tasks, ok := ch.userTasks[c.UserID]; ok {
			delete(tasks, c.ID)
			if len(tasks) == 0 {
				delete(ch.userTasks, c.UserID)
			}
		}
	}
}

func (ch *Channel) GetClientsByDrill(drillID uint) []*Client {
	if ch.Type != ChannelDisplay && ch.Type != ChannelControl {
		return nil
	}

	room := ch.subrooms[drillID]
	result := make([]*Client, 0, len(room))
	for _, c := range room {
		result = append(result, c)
	}
	return result
}

func (ch *Channel) GetClientsByUser(userID uint) []*Client {
	if ch.Type != ChannelTasks {
		return nil
	}

	tasks := ch.userTasks[userID]
	result := make([]*Client, 0, len(tasks))
	for _, c := range tasks {
		result = append(result, c)
	}
	return result
}

func (ch *Channel) ClientCount() int {
	return len(ch.clients)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (m *Manager) HandleDisplay(c *gin.Context) {
	drillIDStr := c.Param("drillId")
	drillID, err := strconv.ParseUint(drillIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的演练ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败 (display): %v", err)
		return
	}

	userID := extractUserID(c)
	clientID := generateClientID()
	client := NewClient(clientID, userID, conn, uint(drillID), ChannelDisplay)

	m.Register <- client

	go client.WritePump()
	go client.ReadPump(m)
}

func (m *Manager) HandleTasks(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败 (tasks): %v", err)
		return
	}

	userID := extractUserID(c)
	clientID := generateClientID()
	client := NewClient(clientID, userID, conn, 0, ChannelTasks)

	m.Register <- client

	go client.WritePump()
	go client.ReadPump(m)
}

func (m *Manager) HandleControl(c *gin.Context) {
	drillIDStr := c.Param("drillId")
	drillID, err := strconv.ParseUint(drillIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的演练ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败 (control): %v", err)
		return
	}

	userID := extractUserID(c)
	clientID := generateClientID()
	client := NewClient(clientID, userID, conn, uint(drillID), ChannelControl)

	m.Register <- client

	go client.WritePump()
	go client.ReadPump(m)
}
