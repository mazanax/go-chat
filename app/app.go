package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
)

type App struct {
	ctx                   context.Context
	UserRepository        db.UserRepository
	AccessTokenRepository db.AccessTokenRepository
	TicketRepository      db.TicketRepository
	OnlineRepository      db.OnlineRepository
	MessageRepository     db.MessageRepository

	Router *mux.Router
}

func New(redisAddr string, redisPassword string, redisDb int) *App {
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
	}

	app.initRoutes()
	return app
}

func (app *App) initRoutes() {
	app.Router.HandleFunc("/", app.IndexHandler()).Methods("GET")
	app.Router.HandleFunc("/api/user", app.UserHandler()).Methods("GET")
	app.Router.HandleFunc("/api/user/{uuid}", app.UserHandler()).Methods("GET")
	app.Router.HandleFunc("/api/users", app.UsersHandler()).Methods("GET")
	app.Router.HandleFunc("/api/signup", app.SignUpHandler()).Methods("POST")
	app.Router.HandleFunc("/api/login", app.LoginHandler()).Methods("POST")
	app.Router.HandleFunc("/api/ticket", app.TicketHandler()).Methods("POST")
	app.Router.HandleFunc("/api/history", app.HistoryHandler()).Methods("GET")
}
