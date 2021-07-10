package models

type AccessToken struct {
	UserID    string
	Token     string
	CreatedAt int
	ExpireAt  int
}

type JsonAccessToken struct {
	Token     string `json:"token"`
	CreatedAt int    `json:"created_at"`
	ExpireAt  int    `json:"expire_at"`
}

type Ticket struct {
	UserID    string
	TokenID   string
	Ticket    string
	CreatedAt int
	ExpireAt  int
}

type JsonTicket struct {
	Ticket    string `json:"ticket"`
	CreatedAt int    `json:"created_at"`
	ExpireAt  int    `json:"expire_at"`
}