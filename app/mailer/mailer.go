package mailer

import (
	"crypto/tls"
	"fmt"
	"github.com/mazanax/go-chat/app/logger"
	"net/mail"
	"net/smtp"
	"strings"
)

type Mail struct {
	to      string
	message []byte
}

type Mailer struct {
	sender string
	host   string
	port   int
	auth   smtp.Auth
	queue  chan Mail
}

func New(authLogin string, senderMail string, senderPassword string, smtpHost string, smtpPort int) *Mailer {
	auth := smtp.PlainAuth("", authLogin, senderPassword, smtpHost)

	return &Mailer{
		auth:   auth,
		host:   smtpHost,
		sender: senderMail,
		port:   smtpPort,

		queue: make(chan Mail, 256),
	}
}

func (mailer *Mailer) Run() {
	for {
		select {
		case message := <-mailer.queue:
			logger.Debug("[mailer] Got message for %s: %s\n", message.to, message.message)

			sender := strings.Trim(mailer.sender, "\n\r")

			client := mailer.getClient()
			if client == nil {
				logger.Error("[mailer] Cannot create client\n")
				continue
			}

			if err := client.Auth(mailer.auth); err != nil {
				logger.Error("[mailer] Auth error: %s\n", err)
				continue
			}

			if err := client.Mail(sender); err != nil {
				logger.Error("[mailer] Cannot set sender: %s\n", err)
				continue
			}

			if err := client.Rcpt(message.to); err != nil {
				logger.Error("[mailer] Cannot set recipient: %s\n", err)
				continue
			}

			w, err := client.Data()
			if err != nil {
				logger.Error("[mailer] Cannot get writer: %s\n", err)
				continue
			}

			if _, err := w.Write(message.message); err != nil {
				logger.Error("[mailer] Cannot write message: %s\n", err)
				_ = w.Close()
				continue
			}

			_ = w.Close()
			_ = client.Quit()
			logger.Debug("[mailer] Sent message %#v\n", message)
		default: // nothing
		}
	}
}

func (mailer *Mailer) getClient() *smtp.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mailer.host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", mailer.host, mailer.port), tlsConfig)
	if err != nil {
		return nil
	}

	client, _ := smtp.NewClient(conn, mailer.host)
	return client
}

func (mailer *Mailer) Enqueue(email string, message string, subject string) {
	from := mail.Address{Name: "MZNX Chat", Address: mailer.sender}
	to := mail.Address{Name: "", Address: email}

	msg := fmt.Sprintf("From: %s\r\n", from.String()) +
		fmt.Sprintf("Reply-To: %s\r\n", from.String()) +
		fmt.Sprintf("To: %s\r\n", to.String()) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		message + "\r\n"

	mailer.queue <- Mail{to: email, message: []byte(msg)}
}
