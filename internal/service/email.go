package service

import (
	"context"
	"fmt"
	"net/smtp"
	"strconv"
)

type PasswordResetEmail struct {
	To        string
	Name      string
	ResetLink string
}

type EmailSender interface {
	SendPasswordReset(ctx context.Context, email PasswordResetEmail) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPEmailSender struct {
	config SMTPConfig
}

func NewSMTPEmailSender(config SMTPConfig) *SMTPEmailSender {
	return &SMTPEmailSender{config: config}
}

func (s *SMTPEmailSender) SendPasswordReset(ctx context.Context, email PasswordResetEmail) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	address := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	body := fmt.Sprintf("To: %s\r\nSubject: Reset your Odessa password\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nHello %s,\r\n\r\nReset your password using this link:\r\n%s\r\n\r\nThis link expires soon. If you did not request this, ignore this email.\r\n", email.To, email.Name, email.ResetLink)
	return smtp.SendMail(address, auth, s.config.From, []string{email.To}, []byte(body))
}
