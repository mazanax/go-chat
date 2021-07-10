package websocket

import (
	"bytes"
	"errors"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	// UUID of user
	userID string
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c

		if err := c.conn.Close(); err != nil {
			logger.Fatal(err.Error())
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Fatal("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		logger.Debug("-> Got new message from %s: %s\n", c.conn.RemoteAddr().String(), string(message))

		_, err = c.hub.messageRepository.StoreMessage(c.userID, models.RegularMessage, string(message))
		if err != nil {
			logger.Error("[websocket] Cannot save message from %s: %s\n", c.userID, err)
			continue
		}
		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	logger.Debug("-> New client: %s\n", c.conn.RemoteAddr().String())
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()

		if err := c.conn.Close(); err != nil {
			logger.Fatal(err.Error())
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	ticketString := r.URL.Query().Get("ticket")
	ticket, err := hub.ticketRepository.GetTicket(ticketString)
	switch {
	case errors.Is(err, db.TicketNotFound):
		logger.Error("[websocket] Ticket %s not found.\n", ticketString)
		_ = conn.Close()
		return
	case err != nil:
		logger.Error(err.Error())
		_ = conn.Close()
		return
	}

	client := &Client{
		userID: ticket.UserID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
	client.hub.register <- client

	err = hub.ticketRepository.RemoveTicket(ticket)
	if err != nil {
		logger.Error("[websocket] Cannot delete ticket %s: %v\n", ticketString, err.Error())
		return
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
