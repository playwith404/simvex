package services

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mailer interface {
	Send(to, subject, body string, isHTML bool) error
}

type SMTPMailer struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPMailer(host string, port int, user, pass, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, user: user, pass: pass, from: from}
}

func (m *SMTPMailer) Send(to, subject, body string, isHTML bool) error {
	if m == nil {
		return fmt.Errorf("mailer not configured")
	}
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}
	headers := []string{
		fmt.Sprintf("From: %s", m.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: %s", contentType),
		"",
	}
	msg := strings.Join(headers, "\r\n") + body

	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}
