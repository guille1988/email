package actions

import (
	"bytes"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/config"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"

	"github.com/go-mail/mail/v2"
)

type SendStress struct {
	dialer          *mail.Dialer
	emailRepository model.Repository
}

func NewSendStress(cfg config.MailConfig, emailRepository model.Repository) *SendStress {
	return &SendStress{
		dialer:          mail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password),
		emailRepository: emailRepository,
	}
}

func (action *SendStress) Execute(to, name, eventID string) error {
	emailRecord := &model.Email{
		EventID: eventID,
		To:      to,
		Subject: "Stress Test — Go App",
		Status:  model.Pending,
		Type:    model.StressEmail,
	}

	if err := action.emailRepository.Create(emailRecord); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		}
		return err
	}

	templatePath := filepath.Join("internal", "domain", "email", "templates", "stress_test.html")
	_, err := os.Stat(templatePath)

	// Fallback for tests running from email/tests/integration/emails
	if os.IsNotExist(err) {
		templatePath = filepath.Join("..", "..", "..", "internal", "domain", "email", "templates", "stress_test.html")
	}

	var tmpl *template.Template
	tmpl, err = template.ParseFiles(templatePath)

	if err != nil {
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	dataStress := dtos.StressEmail{
		Name:  name,
		Email: to,
	}

	err = tmpl.Execute(&body, dataStress)

	if err != nil {
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewMessage()
	message.SetHeader("From", "no-reply@go-app.com")
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Stress Test — Go App")
	message.SetBody("text/html", body.String())

	err = action.dialer.DialAndSend(message)

	if err != nil {
		slog.Error("failed to send stress email", "error", err, "to", to)
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)

		return err
	}

	emailRecord.Body = body.String()
	emailRecord.Status = model.Sent

	return action.emailRepository.Update(emailRecord)
}
