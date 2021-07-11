package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/models"
)

type App struct {
	ctx                          context.Context
	UserRepository               db.UserRepository
	AccessTokenRepository        db.AccessTokenRepository
	TicketRepository             db.TicketRepository
	OnlineRepository             db.OnlineRepository
	MessageRepository            db.MessageRepository
	PasswordResetTokenRepository db.ResetPasswordTokenRepository

	Router *mux.Router

	notifications chan *models.Message
}

func New(redisAddr string, redisPassword string, redisDb int, notifications chan *models.Message) *App {
	ctx := context.Background()
	redisDriver := db.NewRedisDriver(ctx, redisAddr, redisPassword, redisDb)
	app := &App{
		ctx:                   ctx,
		UserRepository:        &redisDriver,
		AccessTokenRepository: &redisDriver,
		TicketRepository:      &redisDriver,
		OnlineRepository:      &redisDriver,
		MessageRepository:     &redisDriver,
		Router:                mux.NewRouter(),
		notifications:         notifications,
	}

	app.initRoutes()
	return app
}

func (app *App) initRoutes() {
	app.Router.HandleFunc("/", app.IndexHandler()).Methods("GET")
	app.Router.HandleFunc("/api/token", app.TokenHandler()).Methods("POST")
	app.Router.HandleFunc("/api/user", app.UserHandler()).Methods("GET")
	app.Router.HandleFunc("/api/user", app.UserHandler()).Methods("PATCH")
	app.Router.HandleFunc("/api/user/{uuid}", app.UserHandler()).Methods("GET")
	app.Router.HandleFunc("/api/users", app.UsersHandler()).Methods("GET")
	app.Router.HandleFunc("/api/online", app.OnlineHandler()).Methods("GET")
	app.Router.HandleFunc("/api/signup", app.SignUpHandler()).Methods("POST")
	app.Router.HandleFunc("/api/login", app.LoginHandler()).Methods("POST")
	app.Router.HandleFunc("/api/logout", app.LogoutHandler()).Methods("POST")
	app.Router.HandleFunc("/api/reset-password", app.ResetPasswordHandler()).Methods("POST")
	app.Router.HandleFunc("/api/ticket", app.TicketHandler()).Methods("POST")
	app.Router.HandleFunc("/api/history", app.HistoryHandler()).Methods("GET")
}
