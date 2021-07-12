package app

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"github.com/mazanax/go-chat/app/requests"
	"github.com/mazanax/go-chat/app/tokens"
	"net/http"
	"path/filepath"
	"time"
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

func (app *App) OnlineHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if err := checkAuthorization(r); errors.Is(err, Unauthorized) {
			logger.Debug("[http] Unauthorized\n")
			sendResponse(w, models.Unauthorized, http.StatusUnauthorized)
			return
		}

		online := app.OnlineRepository.GetOnlineUsers()
		sendResponse(w, online, http.StatusOK)
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

		encryptedPassword, err := app.passwordEncryptor.GenerateHash(req.Password)
		if err != nil {
			logger.Debug("[http] Cannot encode password: %s\n", err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		uuid_, err := app.UserRepository.CreateUser(req.Email, req.Username, req.Name, encryptedPassword)
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

		user, err := app.UserRepository.GetUser(uuid_)
		if err != nil && errors.Is(err, db.UserNotFound) {
			logger.Debug("[http] User #%s not found\n", uuid_)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		app.notifications <- &models.Message{
			ID:        uuid.NewString(),
			Type:      models.UserRegistered,
			Data:      mapUserToJson(user, false),
			CreatedAt: int(time.Now().Unix()),
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
		uuid_ := vars["uuid"]
		if len(uuid_) == 0 && len(accessToken.UserID) > 0 {
			uuid_ = accessToken.UserID
		}

		needEmail := false
		if uuid_ == accessToken.UserID {
			needEmail = true
		}

		user, err := app.UserRepository.GetUser(uuid_)
		if err != nil && errors.Is(err, db.UserNotFound) {
			logger.Debug("[http] User #%s not found\n", uuid_)
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

		if app.passwordEncryptor.CompareHasAndPassword(req.Password, user.Password) {
			logger.Debug("[http] Invalid password %s: %s %s\n", req.Email, r.Method, r.URL)
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

func (app *App) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if err := checkAuthorization(r); errors.Is(err, Unauthorized) {
			logger.Debug("[http] Unauthorized\n")
			sendResponse(w, models.Unauthorized, http.StatusUnauthorized)
			return
		}

		tokenString := parseToken(r)
		accessToken, _ := app.AccessTokenRepository.FindTokenByString(tokenString)

		err := app.AccessTokenRepository.RemoveToken(accessToken)
		if err != nil {
			logger.Debug("[http] Cannot remove access token: %s\n", err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}
		sendResponse(w, nil, http.StatusOK)
	}
}

func (app *App) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)

		req := models.PasswordResetRequest{}
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

		user, err := app.UserRepository.FindUserByEmail(req.Email)
		if err != nil {
			logger.Debug("[http] User %s not found: %s %s\n", req.Email, r.Method, r.URL)
			sendResponse(w, nil, http.StatusOK)
			return
		}

		token, err := tokens.NewPasswordResetToken(app.PasswordResetTokenRepository, &user)
		if err != nil {
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		logger.Debug("[password reset] Code: %s\n", token.Token)
		// here send e-mail

		sendResponse(w, nil, http.StatusOK)
	}
}

func (app *App) TokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)

		req := models.TokenByCodeRequest{}
		err := parse(r, &req)
		if err != nil {
			logger.Error("[http] Cannot parse post body. err=%v\n", err)
			sendResponse(w, models.ErrorResponse{Code: http.StatusBadRequest}, http.StatusBadRequest)
			return
		}

		validationErrors := requests.Validate(req)
		if len(validationErrors) > 0 {
			logger.Debug("[http] Forbidden: %s %s\n", r.Method, r.URL)
			sendResponse(w, nil, http.StatusForbidden)
			return
		}

		token, err := app.PasswordResetTokenRepository.FindTokenByString(req.Code)
		switch {
		case errors.Is(err, db.TokenNotFound):
			logger.Debug("[http] Password reset token %s not found: %s %s\n", req.Code, r.Method, r.URL)
			sendResponse(w, nil, http.StatusForbidden)
			return
		case err != nil:
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		user, err := app.UserRepository.GetUser(token.UserID)
		if err != nil {
			logger.Debug("[http] User %s not found: %s %s\n", req.Code, r.Method, r.URL)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		accessToken, err := tokens.NewToken(app.AccessTokenRepository, &user)
		if err != nil {
			logger.Error("[http] Unexpected error: %s %s %s\n", r.Method, r.URL, err)
			sendResponse(w, models.InternalServerError, http.StatusInternalServerError)
			return
		}

		sendResponse(w, mapAccessTokenToJson(accessToken), http.StatusCreated)
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
