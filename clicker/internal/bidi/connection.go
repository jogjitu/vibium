package bidi

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Connection represents a WebSocket connection.
type Connection struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed bool
}

// Connect establishes a WebSocket connection to the given URL.
func Connect(url string) (*Connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", url, err)
	}

	return &Connection{
		conn: conn,
	}, nil
}

// Send sends a text message over the WebSocket.
func (c *Connection) Send(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("connection closed")
	}

	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Receive receives a text message from the WebSocket.
// Blocks until a message is received.
func (c *Connection) Receive() (string, error) {
	if c.closed {
		return "", fmt.Errorf("connection closed")
	}

	msgType, msg, err := c.conn.ReadMessage()
	if err != nil {
		return "", err
	}

	if msgType != websocket.TextMessage {
		return "", fmt.Errorf("expected text message, got type %d", msgType)
	}

	return string(msg), nil
}

// Close closes the WebSocket connection.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Send close message
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	return c.conn.Close()
}
