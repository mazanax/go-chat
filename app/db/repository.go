package db

import (
	"fmt"
	"github.com/mazanax/go-chat/app/models"
)

var (
	EmailAlreadyExists = fmt.Errorf("user with given email already exists")
	UserNotCreated     = fmt.Errorf("user not created")
	UserNotFound       = fmt.Errorf("user not found")
)

type UserRepository interface {
	IsEmailExists(email string) bool
	CreateUser(email string, name string, encryptedPassword string) (string, error)
	GetUser(id string) (models.User, error)
}
