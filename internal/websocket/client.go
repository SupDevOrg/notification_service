package websocket

import "github.com/gorilla/websocket"

type Client struct {
	Hub    *Hub
	Socket *websocket.Conn
	Send   chan []byte
	UserID uint64
}

func (c *Client) Read() {
	defer func() {
		c.Hub.Unregister <- c
		c.Socket.Close()
	}()

	for {
		if _, _, err := c.Socket.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) Write() {
	defer func() {
        c.Hub.Unregister <- c
        c.Socket.Close()
    }()

	for msg := range c.Send {
		if err := c.Socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
