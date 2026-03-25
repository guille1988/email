package email

import (
	"bytes"
	"email/internal/infrastructure/config"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"

	"github.com/go-mail/mail/v2"
)

type Service struct {
	dialer *mail.Dialer
}

func NewEmailService(cfg config.MailConfig) *Service {
	return &Service{
		dialer: mail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password),
	}
}

func (service *Service) SendWelcomeEmail(to, name string) (string, error) {
	templatePath := filepath.Join("internal", "domain", "email", "templates", "welcome_user.html")
	tmpl, err := template.ParseFiles(templatePath)

	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer

	data := struct {
		Name  string
		Email string
	}{
		Name:  name,
		Email: to,
	}

	err = tmpl.Execute(&body, data)

	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewMessage()
	message.SetHeader("From", "no-reply@go-app.com")
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Bienvenido a Go App")
	message.SetBody("text/html", body.String())

	err = service.dialer.DialAndSend(message)

	if err != nil {
		slog.Error("failed to send email", "error", err, "to", to)
		return body.String(), err
	}

	slog.Info("email sent successfully", "to", to)

	return body.String(), nil
}
