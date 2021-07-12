package config

import (
	"os"
	"strconv"
	"strings"
)

var (
	PublicHost        = os.Getenv("PUBLIC_HOST")
	AllowedOrigins    = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	RedisAddr         = os.Getenv("REDIS_ADDR")
	RedisPassword     = os.Getenv("REDIS_PASSWORD")
	RedisDB, _        = strconv.Atoi(os.Getenv("REDIS_DB"))
	MailerLogin       = os.Getenv("MAILER_LOGIN")
	MailerSender      = os.Getenv("MAILER_SENDER")
	MailerPassword    = os.Getenv("MAILER_PASSWORD")
	MailerSmtpHost    = os.Getenv("MAILER_SMTP_HOST")
	MailerSmtpPort, _ = strconv.Atoi(os.Getenv("MAILER_SMTP_PORT"))
	BCryptCost, _     = strconv.Atoi(os.Getenv("BCRYPT_COST"))
)
