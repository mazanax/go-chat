package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/mazanax/go-chat/app"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"math"
	"time"
)

const (
	DurationInHours = 48
)

func NewToken(app *app.App, user *models.User) (models.AccessToken, error) {
	randomString := randomHexString(64)
	tokenUUID, err := app.AccessTokenRepository.CreateToken(user, randomString, time.Duration(DurationInHours)*time.Hour)
	if err != nil {
		return models.AccessToken{}, err
	}

	token, err := app.AccessTokenRepository.GetToken(tokenUUID)
	if err != nil && errors.Is(err, db.TokenNotFound) {
		return token, err
	}

	return token, nil
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
