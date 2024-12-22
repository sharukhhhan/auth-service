package sender

type Email interface {
	SendWarningEmail(toEmail, subject, body string) error
	EnsureSMTPConnection() error
}

type Sender struct {
	Email
}

func NewSender(email Email) *Sender {
	return &Sender{
		Email: email,
	}
}
