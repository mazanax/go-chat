package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"math"
	"net/http"
)

func randomHexString(length int) string {
	buff := make([]byte, int(math.Ceil(float64(length)/2)))
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:length]
}

func parse(r *http.Request, data interface{}) error {
	return json.NewDecoder(r.Body).Decode(data)
}

func parseToken(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 {
		tokenString = tokenString[7:]
	}

	return tokenString
}

func sendResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Error("Cannot format json. err=%v\n", err)
	}
}

func mapUserToJson(user models.User, withEmail bool) models.JsonUser {
	email := user.Email
	if !withEmail {
		email = ""
	}

	return models.JsonUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func mapAccessTokenToJson(token models.AccessToken) models.JsonAccessToken {
	return models.JsonAccessToken{
		Token:     token.Token,
		CreatedAt: token.CreatedAt,
		ExpireAt:  token.ExpireAt,
	}
}
