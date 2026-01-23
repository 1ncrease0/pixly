package email

import (
	"fmt"
	"github.com/1ncrease0/pixly/services/notification/internal/domain"
	"gopkg.in/gomail.v2"
)

type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	address  string
}

func NewSender(host string, port int, username string, password string, from string, address string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		address:  address,
	}
}

func (s *Sender) SendVerification(email domain.Email, code string) error {
	dialer := gomail.NewDialer(s.host, s.port, s.username, s.password)
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", email.String())
	m.SetHeader("Subject", "Please verify your email address")

	link := s.address + code
	m.SetBody("text/plain", fmt.Sprintf("Verification link: %s", link))
	if err := dialer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
