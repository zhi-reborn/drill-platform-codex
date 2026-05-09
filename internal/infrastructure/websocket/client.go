package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	UserID   uint
	Conn     *websocket.Conn
	Send     chan []byte
	DrillID  uint
	Type     string
	Mu       sync.Mutex
	lastPong time.Time
}

func NewClient(id string, userID uint, conn *websocket.Conn, drillID uint, connType string) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		DrillID:  drillID,
		Type:     connType,
		lastPong: time.Now(),
	}
}

func (c *Client) ReadPump(m *Manager) {
	defer func() {
		m.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		c.lastPong = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("WebSocket 读取错误 [%s]: %v", c.ID, err)
			}
			break
		}

		if len(message) == 0 {
			continue
		}

		m.handleClientMessage(c, message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return ErrSendChannelFull
	}
}

func (c *Client) Close() {
	c.Conn.Close()
	close(c.Send)
}
