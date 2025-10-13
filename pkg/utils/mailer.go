package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host      string
	Port      int
	User      string
	Pass      string
	FromEmail string
	FromName  string
}

type Mailer struct{ cfg SMTPConfig }

func NewMailer(cfg SMTPConfig) *Mailer { return &Mailer{cfg: cfg} }

func (m *Mailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	from := m.cfg.FromEmail
	header := []string{
		fmt.Sprintf("From: %s <%s>", m.cfg.FromName, from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	msg := strings.Join(header, "\r\n") + "\r\r" + body
	host, _, _ := net.SplitHostPort(addr)
	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, host)
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
		conf := &tls.Config{ServerName: host}
		if err := c.StartTLS(conf); err != nil {
			return err
		}
	}
	if m.cfg.User != "" {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}
