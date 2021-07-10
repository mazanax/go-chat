package app

import (
	"encoding/json"
	"fmt"
	"github.com/mazanax/go-chat/app/logger"
	"net/http"
)

var Unauthorized = fmt.Errorf("unathorized")

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

func checkAuthorization(r *http.Request) error {
	token := parseToken(r)
	if len(token) == 0 {
		return Unauthorized
	}

	return nil
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
