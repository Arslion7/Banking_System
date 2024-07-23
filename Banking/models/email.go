package models

import (
	"fmt"

	"github.com/go-mail/mail/v2"
)

const DefaultSender = "aaa@gmail.com"

type EmailService struct {
	DefaultSender string
	dialer *mail.Dialer
}

type Email struct {
	From string
	To string
	Subject string
	Plaintext string
	HTML string
}

type SMTPConfig struct {
	Host string
	Port int
	Username string
	Password string
}

func NewMessageService (config SMTPConfig) *EmailService {
	es := EmailService{
		dialer: mail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}
	return &es
}

func (es *EmailService) Send(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("To", email.To)
	es.setFrom(msg, email)
	msg.SetHeader("Subject", email.Subject)
	if email.Plaintext != "" {
		msg.SetBody("text/plain", email.Plaintext)
	}
	if email.HTML != "" {
		msg.AddAlternative("text/html", email.HTML)
	}
	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send message %w", err)
	}
	return nil
}

func (es *EmailService) setFrom(msg *mail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.SetHeader("From", from)
}

func (es *EmailService) ForgotPassword(to, resetURL string) error {
	email := Email{
		To: to,
		Subject: "Reset your password",
		Plaintext: "To reset your password, please visit the following link: " + resetURL,
		HTML: `<p>To reset your password, please visit the following link: <a href="` + resetURL + `">` + resetURL + `</a></p>`,
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot password %w", err)
	}
	return nil
}