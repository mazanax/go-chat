package websocket

import (
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
)

type Hub struct {
	ticketRepository  db.TicketRepository
	onlineRepository  db.OnlineRepository
	messageRepository db.MessageRepository

	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub(
	ticketRepository db.TicketRepository,
	onlineRepository db.OnlineRepository,
	messageRepository db.MessageRepository,
) *Hub {
	return &Hub{
		ticketRepository:  ticketRepository,
		onlineRepository:  onlineRepository,
		messageRepository: messageRepository,

		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
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
