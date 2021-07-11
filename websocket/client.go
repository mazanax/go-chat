package websocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
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
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("origin") == "http://localhost:3000" || r.Header.Get("origin") == "https://localhost:3000"
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	// UUID of user
	userID string
	hub    *Hub
	conn   *websocket.Conn
	send   chan *models.Message
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.hub.notifications <- &models.Message{
			ID:        uuid.NewString(),
			UserID:    c.userID,
			CreatedAt: int(time.Now().Unix()),
			Type:      models.UserDisconnected,
		}

		if err := c.conn.Close(); err != nil {
			logger.Error(err.Error())
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

		msg := models.WebsocketMessage{}
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Error("[websocket] Cannot decode message: %s\n", err)
			continue
		}

		messageID, err := c.hub.messageRepository.StoreMessage(c.userID, models.RegularMessage, msg.ID, msg.Text)
		if err != nil {
			logger.Error("[websocket] Cannot save message from %s: %s\n", c.userID, err)
			continue
		}
		messageModel, err := c.hub.messageRepository.GetMessage(messageID)
		if err != nil {
			logger.Error("[websocket] Cannot get message #%s from %s: %s\n", messageID, c.userID, err)
			continue
		}

		c.hub.broadcast <- &messageModel
	}
}

func (c *Client) writePump() {
	logger.Debug("-> New client: %s\n", c.conn.RemoteAddr().String())
	ticker := time.NewTicker(pingPeriod)

	c.hub.notifications <- &models.Message{
		ID:        uuid.NewString(),
		UserID:    c.userID,
		CreatedAt: int(time.Now().Unix()),
		Type:      models.UserConnected,
	}

	defer func() {
		ticker.Stop()

		if err := c.conn.Close(); err != nil {
			logger.Debug("User %s closed connection: %s\n", c.conn.RemoteAddr().String(), err.Error())
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
			message_ := encodeMessage(message)
			if len(message_) > 0 {
				_, _ = w.Write(encodeMessage(message))
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				message_ := encodeMessage(<-c.send)
				if len(message_) > 0 {
					_, _ = w.Write(newline)
					_, _ = w.Write(encodeMessage(message))
				}
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
	logger.Debug("[websocket] Incoming connection\n")
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
		send:   make(chan *models.Message, 256),
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

func encodeMessage(message *models.Message) []byte {
	jsonMessage, err := json.Marshal(mapMessageToJson(*message))
	if err != nil {
		logger.Error("[websocket] Cannot encode message: %s\n", err)
		return []byte{}
	}

	return jsonMessage
}

func mapMessageToJson(message models.Message) models.JsonMessage {
	return models.JsonMessage{
		ID:        message.ID,
		UserID:    message.UserID,
		Type:      message.Type,
		CreatedAt: message.CreatedAt,
		Text:      message.Text,
		Data:      message.Data,
	}
}
