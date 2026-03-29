package actions

import (
	"bytes"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/config"
	"email/internal/shared/messaging/rabbitmq/dtos"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-mail/mail/v2"
)

type SendWelcome struct {
	dialer          *mail.Dialer
	emailRepository model.Repository
}

func NewSendWelcome(cfg config.MailConfig, emailRepository model.Repository) *SendWelcome {
	return &SendWelcome{
		dialer:          mail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password),
		emailRepository: emailRepository,
	}
}

func (action *SendWelcome) Execute(to, name string) error {
	emailRecord := &model.Email{
		To:      to,
		Subject: "Bienvenido a Go App",
		Status:  model.Pending,
	}

	if err := action.emailRepository.Create(emailRecord); err != nil {
		return err
	}

	templatePath := filepath.Join("internal", "domain", "email", "templates", "welcome_user.html")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Fallback for tests running from email/tests/integration/emails
		templatePath = filepath.Join("..", "..", "..", "internal", "domain", "email", "templates", "welcome_user.html")
	}
	tmpl, err := template.ParseFiles(templatePath)

	if err != nil {
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	dataWelcome := dtos.WelcomeEmail{
		Name:  name,
		Email: to,
	}

	err = tmpl.Execute(&body, dataWelcome)

	if err != nil {
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewMessage()
	message.SetHeader("From", "no-reply@go-app.com")
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Bienvenido a Go App")
	message.SetBody("text/html", body.String())

	err = action.dialer.DialAndSend(message)

	if err != nil {
		slog.Error("failed to send email", "error", err, "to", to)
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)

		return err
	}

	emailRecord.Body = body.String()
	emailRecord.Status = model.Sent

	return action.emailRepository.Update(emailRecord)
}
