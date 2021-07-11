package websocket

import (
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
)

type Hub struct {
	ticketRepository  db.TicketRepository
	onlineRepository  db.OnlineRepository
	messageRepository db.MessageRepository

	// this channel is used to send notifications from the REST API
	notifications chan *models.Message

	clients    map[*Client]bool
	broadcast  chan *models.Message
	register   chan *Client
	unregister chan *Client
}

func NewHub(
	ticketRepository db.TicketRepository,
	onlineRepository db.OnlineRepository,
	messageRepository db.MessageRepository,
	notifications chan *models.Message,
) *Hub {
	return &Hub{
		ticketRepository:  ticketRepository,
		onlineRepository:  onlineRepository,
		messageRepository: messageRepository,

		notifications: notifications,

		broadcast:  make(chan *models.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case notification := <-h.notifications:
			logger.Debug("[websocket] Received new notification: %v\n", notification)
			for client := range h.clients {
				select {
				case client.send <- notification:
				default:
					close(client.send)
					delete(h.clients, client)

					err := h.onlineRepository.RemoveUserOnline(client.userID)
					if err != nil {
						logger.Fatal("[websocket] Cannot remove online user: %v\n", err)
					}
				}
			}
		case client := <-h.register:
			logger.Debug("[websocket] User connected\n")
			err := h.onlineRepository.CreateUserOnline(client.userID)
			if err != nil {
				logger.Fatal("[websocket] Cannot save online user: %v\n", err)
			}

			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				err := h.onlineRepository.RemoveUserOnline(client.userID)
				if err != nil {
					logger.Fatal("[websocket] Cannot remove online user: %v\n", err)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)

					err := h.onlineRepository.RemoveUserOnline(client.userID)
					if err != nil {
						logger.Fatal("[websocket] Cannot remove online user: %v\n", err)
					}
				}
			}
		}
	}
}
