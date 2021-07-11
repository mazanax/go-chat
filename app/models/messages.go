package models

const (
	UserRegistered   = -1
	UserConnected    = -100
	UserDisconnected = -101
	RegularMessage   = 0
)

type WebsocketMessage struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Message struct {
	ID        string
	UserID    string
	Type      int
	CreatedAt int
	Text      string
	Data      interface{}
}

type JsonMessage struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      int         `json:"type"`
	CreatedAt int         `json:"created_at"`
	Text      string      `json:"text"`
	Data      interface{} `json:"data"`
}
