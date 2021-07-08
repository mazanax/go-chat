package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
)

type App struct {
	ctx            context.Context
	UserRepository db.UserRepository
	Router         *mux.Router
}

func New(redisAddr string, redisPassword string, redisDb int) *App {
	ctx := context.Background()
	redisDriver := db.NewRedisDriver(ctx, redisAddr, redisPassword, redisDb)
	app := &App{
		ctx:            ctx,
		UserRepository: &redisDriver,
		Router:         mux.NewRouter(),
	}

	app.initRoutes()
	return app
}

func (app *App) initRoutes() {
	app.Router.HandleFunc("/", app.IndexHandler()).Methods("GET")
	app.Router.HandleFunc("/api/user/{uuid}", app.UserHandler()).Methods("GET")
	app.Router.HandleFunc("/api/users", app.UsersHandler()).Methods("GET")
	app.Router.HandleFunc("/api/signup", app.SignUpHandler()).Methods("POST")
	app.Router.HandleFunc("/api/login", nil).Methods("POST")
	app.Router.HandleFunc("/api/history", nil).Methods("GET")
}
