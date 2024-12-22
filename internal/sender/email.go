package sender

import (
	"fmt"
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

func NewEmailSender(host string, port int, user, password string) *EmailSender {
	return &EmailSender{
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPUser:     user,
		SMTPPassword: password,
	}
}

func (e *EmailSender) SendWarningEmail(toEmail, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", e.SMTPUser)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	dialer := gomail.NewDialer(e.SMTPHost, e.SMTPPort, e.SMTPUser, e.SMTPPassword)

	return dialer.DialAndSend(message)
}

func (e *EmailSender) EnsureSMTPConnection() error {
	dialer := gomail.NewDialer(e.SMTPHost, e.SMTPPort, e.SMTPUser, e.SMTPPassword)

	conn, err := dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	return nil
}
