package models

const (
	UserRegistered   = -1
	UserConnected    = -100
	UserDisconnected = -101
	RegularMessage   = 0
	UpdateMessage    = 1
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
	UpdatedAt int
	Text      string
	Data      interface{}
}

type JsonMessage struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      int         `json:"type"`
	CreatedAt int         `json:"created_at"`
	UpdatedAt int         `json:"updated_at"`
	Text      string      `json:"text"`
	Data      interface{} `json:"data"`
}

type UpdateMessageRequest struct {
	Text string `json:"text" validate:"min=1,max=2000000"` // rough rounded limit 2 MiB
}
