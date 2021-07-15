package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/mazanax/go-chat/app/db"
	"github.com/mazanax/go-chat/app/mailer"
	"github.com/mazanax/go-chat/app/models"
	"github.com/mazanax/go-chat/app/security"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	MailerLogin    string
	MailerSender   string
	MailerPassword string
	MailerSmtpHost string
	MailerSmtpPort int

	BCryptCost int
}

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

func New(config Config, notifications chan *models.Message) *App {
	ctx := context.Background()
	redisDriver := db.NewRedisDriver(ctx, config.RedisAddr, config.RedisPassword, config.RedisDB)
	bcryptEncryptor := security.NewBcryptEncryptor(config.BCryptCost)
	mailer_ := mailer.New(
		config.MailerLogin,
		config.MailerSender,
		config.MailerPassword,
		config.MailerSmtpHost,
		config.MailerSmtpPort,
	)

	app := &App{
		ctx:                          ctx,
		UserRepository:               redisDriver,
		AccessTokenRepository:        redisDriver,
		TicketRepository:             redisDriver,
		OnlineRepository:             redisDriver,
		MessageRepository:            redisDriver,
		PasswordResetTokenRepository: redisDriver,

		Router:            mux.NewRouter(),
		Mailer:            mailer_,
		passwordEncryptor: bcryptEncryptor,
		notifications:     notifications,
	}

	app.initRoutes()
	return app
}

func (app *App) initRoutes() {
	// GET or PATCH user
	app.Router.HandleFunc("/api/user", app.UserHandler()).Methods("GET", "PATCH")
	// GET user info
	app.Router.HandleFunc("/api/user/{uuid}", app.UserHandler()).Methods("GET")
	// GET users list
	app.Router.HandleFunc("/api/users", app.UsersHandler()).Methods("GET")
	// GET list of UUID online users
	app.Router.HandleFunc("/api/online", app.OnlineHandler()).Methods("GET")
	// Create new user
	app.Router.HandleFunc("/api/signup", app.SignUpHandler()).Methods("POST")
	// Generate access token with Email and Password
	app.Router.HandleFunc("/api/login", app.LoginHandler()).Methods("POST")
	// Generate access token with PasswordResetToken
	app.Router.HandleFunc("/api/token", app.TokenHandler()).Methods("POST")
	// Invalidate current access token
	app.Router.HandleFunc("/api/logout", app.LogoutHandler()).Methods("POST")
	// Generate PasswordResetToken
	app.Router.HandleFunc("/api/reset-password", app.ResetPasswordHandler()).Methods("POST")
	// Generate ticket for websocket authentication
	app.Router.HandleFunc("/api/ticket", app.TicketHandler()).Methods("POST")
	// GET last 100 messages
	app.Router.HandleFunc("/api/history", app.HistoryHandler()).Methods("GET")
	// PATCH message
	app.Router.HandleFunc("/api/message/{uuid}", app.UpdateMessageHandler()).Methods("PATCH")
}
