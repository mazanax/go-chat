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
