package db

import (
	"fmt"
	"github.com/mazanax/go-chat/app/models"
	"time"
)

var (
	EmailAlreadyExists    = fmt.Errorf("user with given email already exists")
	UsernameAlreadyExists = fmt.Errorf("user with given username already exists")
	UserNotCreated        = fmt.Errorf("user not created")
	UserNotFound          = fmt.Errorf("user not found")
	TokenNotCreated       = fmt.Errorf("token not created")
	TokenNotFound         = fmt.Errorf("token not found")
	TicketNotFound        = fmt.Errorf("ticket not found")
	MessageNotFound       = fmt.Errorf("message not found")
)

type UserRepository interface {
	IsEmailExists(email string) bool
	IsUsernameExists(username string) bool
	CreateUser(email string, username string, name string, encryptedPassword string) (string, error)
	GetUser(id string) (models.User, error)
	GetUsers() []models.User
	FindUserByEmail(email string) (models.User, error)
	UpdateUserField(user *models.User, field string, value string) error
}

type TicketRepository interface {
	CreateTicket(token *models.AccessToken, randomString string, duration time.Duration) error
	GetTicket(ticket string) (models.Ticket, error)
	RemoveTicket(ticket models.Ticket) error
}

type OnlineRepository interface {
	GetOnlineUsers() []string
	CreateUserOnline(userUUID string) error
	RemoveUserOnline(userUUID string) error
}

type MessageRepository interface {
	StoreMessage(userID string, messageType int, messageUUID string, text string) (string, error)
	UpdateMessage(message *models.Message, newText string) (models.Message, error)
	GetMessage(id string) (models.Message, error)
	GetMessages(count int) []models.Message
}

type AccessTokenRepository interface {
	CreateToken(user *models.User, randomString string, duration time.Duration) (string, error)
	GetToken(id string) (models.AccessToken, error)
	FindTokenByString(token string) (models.AccessToken, error)
	RemoveToken(token models.AccessToken) error
}

type ResetPasswordTokenRepository interface {
	CreateResetPasswordToken(user *models.User, randomString string, duration time.Duration) (string, error)
	GetResetPasswordToken(id string) (models.PasswordResetToken, error)
	FindResetPasswordTokenByUser(user *models.User) (models.PasswordResetToken, error)
	FindResetPasswordTokenByString(token string) (models.PasswordResetToken, error)
	RemoveResetPasswordToken(token models.PasswordResetToken) error
}
