package main

import (
	"flag"
	"fmt"
	"github.com/mazanax/go-chat/app"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"github.com/mazanax/go-chat/websocket"
	"github.com/rs/cors"
	"net/http"
	"os"
)

func main() {
	logger.Debug("[Go Chat v0.0.1]\n")
	host := flag.String("host", "<none>", "Host to listen to")
	port := flag.Int("port", -1, "Port to listen to")
	flag.Parse()

	if *host == "<none>" || *port <= 0 {
		logger.Debug("Usage:\n")
		logger.Debug("    chat -host=<HOST> -port=<PORT>\n")
		os.Exit(1)
	}
	logger.Debug("Starting listen to %s:%d...\n", *host, *port)

	notifications := make(chan *models.Message)

	app_ := app.New("0.0.0.0:6379", "", 0, notifications)
	hub := websocket.NewHub(app_.TicketRepository, app_.OnlineRepository, app_.MessageRepository, notifications)
	go hub.Run()
	app_.Router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[http] Incoming Websocket connection\n")
		websocket.ServeWs(hub, w, r)
	})
	http.HandleFunc("/", app_.Router.ServeHTTP)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), c.Handler(app_.Router))
	if err != nil {
		logger.Error("ListenAndServe: %s\n", err.Error())
		os.Exit(2)
	}
}
