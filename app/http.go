package app

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"github.com/mazanax/go-chat/app/requests"
	"github.com/mazanax/go-chat/app/tokens"
	"net/http"
	"path/filepath"
)

// region HttpHandlers

func (app *App) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		// path to file should be from binary location
		http.ServeFile(w, r, filepath.Join("html", "index.html"))
	}
}

func (app *App) UsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if err := checkAuthorization(r); errors.Is(err, Unauthorized) {
			logger.Debug("[http] Unauthorized\n")
			sendResponse(w, models.Unauthorized, http.StatusUnauthorized)
			return
		}

		tokenString := parseToken(r)
		accessToken, _ := app.AccessTokenRepository.FindTokenByString(tokenString)

		users := app.UserRepository.GetUsers()
		var jsonUsers []models.JsonUser
		for _, user := range users {
			withEmail := accessToken.UserID == user.ID
			jsonUsers = append(jsonUsers, mapUserToJson(user, withEmail))
		}

		sendResponse(w, jsonUsers, http.StatusOK)
	}
}

func (app *App) SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		req := models.CreateUserRequest{}
		err := parse(r, &req)
		if err != nil {
			logger.Error("[http] Cannot parse post body. err=%v\n", err)
			sendResponse(w, models.ErrorResponse{Code: http.StatusBadRequest}, http.StatusBadRequest)
			return
		}

		validationErrors := requests.Validate(req)
		if len(validationErrors) > 0 {
			logger.Debug("[http] Bad request: %s %s\n", r.Method, r.URL)
			sendResponse(w, nil, http.StatusBadRequest)
			return
		}

		uuid, err := app.UserRepository.CreateUser(req.Email, req.Username, req.Name, req.Password)
		switch {
		case errors.Is(err, db.EmailAlreadyExists):
			logger.Debug("[http] User with email %s already exists: %s %s\n", req.Email, r.Method, r.URL)
			sendResponse(w, models.EmailAlreadyExists, http.StatusBadRequest)
			return
		case errors.Is(err, db.UsernameAlreadyExists):
			logger.Debug("[http] User with username %s already exists: %s %s\n", req.Username, r.Method, r.URL)
			sendResponse(w, models.UsernameAlreadyExists, http.StatusBadRequest)
			return
		case err != nil:
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		user, err := app.UserRepository.GetUser(uuid)
		if err != nil && errors.Is(err, db.UserNotFound) {
			logger.Debug("[http] User #%s not found\n", uuid)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		sendResponse(w, mapUserToJson(user, true), http.StatusCreated)
	}
}

func (app *App) UserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if err := checkAuthorization(r); errors.Is(err, Unauthorized) {
			logger.Debug("[http] Unauthorized\n")
			sendResponse(w, models.Unauthorized, http.StatusUnauthorized)
			return
		}

		tokenString := parseToken(r)
		accessToken, err := app.AccessTokenRepository.FindTokenByString(tokenString)
		if err != nil {
			accessToken = models.AccessToken{}
		}

		vars := mux.Vars(r)
		uuid := vars["uuid"]
		if len(uuid) == 0 && len(accessToken.UserID) > 0 {
			uuid = accessToken.UserID
		}

		needEmail := false
		if uuid == accessToken.UserID {
			needEmail = true
		}

		user, err := app.UserRepository.GetUser(uuid)
		if err != nil && errors.Is(err, db.UserNotFound) {
			logger.Debug("[http] User #%s not found\n", uuid)
			sendResponse(w, models.UserNotFound, http.StatusNotFound)
			return
		}

		sendResponse(w, mapUserToJson(user, needEmail), http.StatusOK)
	}
}

func (app *App) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)

		req := models.LoginRequest{}
		err := parse(r, &req)
		if err != nil {
			logger.Error("[http] Cannot parse post body. err=%v\n", err)
			sendResponse(w, models.ErrorResponse{Code: http.StatusBadRequest}, http.StatusBadRequest)
			return
		}

		validationErrors := requests.Validate(req)
		if len(validationErrors) > 0 {
			logger.Debug("[http] Bad request: %s %s\n", r.Method, r.URL)
			sendResponse(w, nil, http.StatusUnauthorized)
			return
		}

		user, err := app.UserRepository.FindUserByEmail(req.Email)
		if err != nil {
			logger.Debug("[http] User %s not found: %s %s\n", req.Email, r.Method, r.URL)
			sendResponse(w, models.InvalidCredentials, http.StatusUnauthorized)
			return
		}

		token, err := tokens.NewToken(app.AccessTokenRepository, &user)
		if err != nil {
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		sendResponse(w, mapAccessTokenToJson(token), http.StatusOK)
	}
}

func (app *App) TicketHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if err := checkAuthorization(r); errors.Is(err, Unauthorized) {
			logger.Debug("[http] Unauthorized\n")
			sendResponse(w, models.Unauthorized, http.StatusUnauthorized)
			return
		}

		tokenString := parseToken(r)
		accessToken, _ := app.AccessTokenRepository.FindTokenByString(tokenString)

		ticket, err := tokens.NewTicket(app.TicketRepository, &accessToken)
		if err != nil {
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		sendResponse(w, mapTicketToJson(ticket), http.StatusCreated)
	}
}

func (app *App) HistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)

		var jsonMessages []models.JsonMessage
		messages := app.MessageRepository.GetMessages(100)

		for _, message := range messages {
			jsonMessages = append(jsonMessages, mapMessageToJson(message))
		}

		sendResponse(w, jsonMessages, http.StatusOK)
	}
}

// endregion
