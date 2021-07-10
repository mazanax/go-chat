package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"math"
	"time"
)

const (
	TokenDurationHours    = 48
	TicketDurationSeconds = 45
)

func NewToken(accessTokenRepository db.AccessTokenRepository, user *models.User) (models.AccessToken, error) {
	randomString := randomHexString(64)
	tokenUUID, err := accessTokenRepository.CreateToken(user, randomString, time.Duration(TokenDurationHours)*time.Hour)
	if err != nil {
		return models.AccessToken{}, err
	}

	token, err := accessTokenRepository.GetToken(tokenUUID)
	if err != nil && errors.Is(err, db.TokenNotFound) {
		return token, err
	}

	return token, nil
}

func NewTicket(ticketRepository db.TicketRepository, accessToken *models.AccessToken) (models.Ticket, error) {
	randomString := randomHexString(32)
	err := ticketRepository.CreateTicket(accessToken, randomString, time.Duration(TicketDurationSeconds)*time.Second)
	if err != nil {
		return models.Ticket{}, err
	}

	return ticketRepository.GetTicket(randomString)
}

func randomHexString(length int) string {
	buff := make([]byte, int(math.Ceil(float64(length)/2)))
	_, err := rand.Read(buff)
	if err != nil {
		logger.Fatal("[TokenGenerator] Unexpected error: %s\n", err.Error())
	}

	str := hex.EncodeToString(buff)
	return str[:length]
}
