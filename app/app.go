package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/mailer"
	"github.com/mazanax/go-chat/app/models"
	"github.com/mazanax/go-chat/app/security"
)

type App struct {
	ctx                          context.Context
	UserRepository               db.UserRepository
	AccessTokenRepository        db.AccessTokenRepository
	TicketRepository             db.TicketRepository
	OnlineRepository             db.OnlineRepository
	MessageRepository            db.MessageRepository
	PasswordResetTokenRepository db.ResetPasswordTokenRepository

	Router            *mux.Router
	Mailer            *mailer.Mailer
	passwordEncryptor security.PasswordEncryptor

	notifications chan *models.Message
}

func New(redisAddr string, redisPassword string, redisDb int, notifications chan *models.Message) *App {
	ctx := context.Background()
	redisDriver := db.NewRedisDriver(ctx, redisAddr, redisPassword, redisDb)
	bcryptEncryptor := security.NewBcryptEncryptor(14)
	mailer_ := mailer.New("durov", "durov@yandex.ru", "vmolhgmlxdgluhpi", "smtp.yandex.ru", 465)

	app := &App{
		ctx:                          ctx,
		UserRepository:               &redisDriver,
		AccessTokenRepository:        &redisDriver,
		TicketRepository:             &redisDriver,
		OnlineRepository:             &redisDriver,
		MessageRepository:            &redisDriver,
		PasswordResetTokenRepository: &redisDriver,

		Router:            mux.NewRouter(),
		Mailer:            &mailer_,
		passwordEncryptor: &bcryptEncryptor,
		notifications:     notifications,
	}

	app.initRoutes()
	return app
}

func (app *App) initRoutes() {
	app.Router.HandleFunc("/", app.IndexHandler()).Methods("GET")
	app.Router.HandleFunc("/api/token", app.TokenHandler()).Methods("POST")
	app.Router.HandleFunc("/api/user", app.UserHandler()).Methods("GET", "PATCH")
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
