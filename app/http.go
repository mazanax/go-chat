package app

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

func (app *App) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		http.ServeFile(w, r, "../html/index.html")
	}
}

func (app *App) UsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)
		if r.URL.Path != "/users" {
			logger.Debug("[http] Not found: %s\n", r.URL)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// here return connected users
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

		v := validator.New()
		err = v.Struct(req)

		var validationErrors []string
		if err != nil {
			for _, e := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, e.Translate(nil))
			}
		}

		if len(validationErrors) > 0 {
			logger.Debug("[http] Bad request: %s %s\n", r.Method, r.URL)
			sendResponse(w, nil, http.StatusBadRequest)

			//http.Error(w, strings.Join(errors, "\n"), http.StatusBadRequest)
			return
		}

		uuid, err := app.UserRepository.CreateUser(req.Email, req.Name, req.Password)
		switch {
		case err == db.EmailAlreadyExists:
			logger.Debug("[http] User with email %s already exists: %s %s\n", req.Email, r.Method, r.URL)
			sendResponse(w, models.UserAlreadyExists, http.StatusBadRequest)
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

		sendResponse(w, mapUserToJson(user), http.StatusCreated)
	}
}

func (app *App) UserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Request URL: %s %s\n", r.Method, r.URL)

		vars := mux.Vars(r)
		uuid := vars["uuid"]
		user, err := app.UserRepository.GetUser(uuid)

		if err != nil && errors.Is(err, db.UserNotFound) {
			logger.Debug("[http] User #%s not found\n", uuid)
			sendResponse(w, models.UserNotFound, http.StatusNotFound)
			return
		}

		sendResponse(w, mapUserToJson(user), http.StatusOK)
	}
}

func ServeSignUp() {}
