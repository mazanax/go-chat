package models

const (
	RegularMessage = iota
	NewDay
)

type Message struct {
	ID        string
	UserID    string
	Type      int
	CreatedAt int
	Text      string
}

type JsonMessage struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      int    `json:"type"`
	CreatedAt int    `json:"created_at"`
	Text      string `json:"text"`
}
