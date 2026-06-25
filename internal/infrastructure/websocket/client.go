package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"drill-platform/internal/infrastructure/events"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        string
	UserID    uint
	Conn      *websocket.Conn
	Send      chan events.WSMessage
	DrillID   uint
	Type      string
	Mu        sync.Mutex
	lastPong  time.Time
	closeOnce sync.Once
	sendOnce  sync.Once
	closed    bool
}

func NewClient(id string, userID uint, conn *websocket.Conn, drillID uint, connType string) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan events.WSMessage, 256),
		DrillID:  drillID,
		Type:     connType,
		lastPong: time.Now(),
	}
}

func (c *Client) ReadPump(m *Manager) {
	defer func() {
		m.Unregister <- c
		c.Close()
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

// WritePump reads WSMessages from Send and writes them to the WebSocket
// connection. When multiple messages are queued, they are batched into a
// single JSON array ([msg1,msg2,...]) via json.Marshal([]WSMessage) so the
// client receives one parseable frame instead of newline-concatenated JSON.
func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			batch := []events.WSMessage{msg}
			// Drain any additional queued messages so they share one frame.
			for len(batch) < 64 {
				select {
				case extra, ok := <-c.Send:
					if !ok {
						break
					}
					batch = append(batch, extra)
				default:
					goto write
				}
			}
		write:
			data, err := json.Marshal(batch)
			if err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
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

// SendJSON enqueues a WSMessage for delivery to this client. It is safe to
// call from any goroutine; the recover guards against the rare race where
// the hub closes Send between the closed check and the channel send.
func (c *Client) SendJSON(msg events.WSMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrClientClosed
		}
	}()

	c.Mu.Lock()
	if c.closed {
		c.Mu.Unlock()
		return ErrClientClosed
	}
	c.Mu.Unlock()

	select {
	case c.Send <- msg:
		return nil
	default:
		return ErrSendChannelFull
	}
}

// Close marks the client as closed. It does NOT close the Send channel —
// only the hub goroutine (processing an Unregister job) closes Send via
// closeSend, so writes and closure are serialized and never race.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.Mu.Lock()
		c.closed = true
		c.Mu.Unlock()
	})
}

// closeSend closes the Send channel. Called only by the hub goroutine when
// processing an Unregister job (or evicting a slow client). The sendOnce
// guard prevents double-close if the hub processes the same client twice.
func (c *Client) closeSend() {
	c.sendOnce.Do(func() {
		close(c.Send)
	})
}
