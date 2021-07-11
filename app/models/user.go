package models

type User struct {
	ID        string
	Email     string
	Username  string
	Name      string
	Password  string
	CreatedAt int
	UpdatedAt int
}

type JsonUser struct {
	ID        string `json:"id"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30,slug"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Password string `json:"password" validate:"required,min=6,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=255"`
}
