package main

import (
	"flag"
	"fmt"
	"github.com/mazanax/go-chat/app"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/websocket"
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

	app_ := app.New("0.0.0.0:6379", "", 0)

	hub := websocket.NewHub(app_.TicketRepository, app_.OnlineRepository, app_.MessageRepository)
	go hub.Run()
	http.HandleFunc("/", app_.Router.ServeHTTP)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
	if err != nil {
		logger.Error("ListenAndServe: %s\n", err.Error())
		os.Exit(2)
	}
}
