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

type SendWelcome struct {
	dialer          *mail.Dialer
	emailRepository model.Repository
	template        *template.Template
}

func NewSendWelcome(cfg config.MailConfig, emailRepository model.Repository) *SendWelcome {
	primaryPath := filepath.Join("internal", "domain", "email", "templates", "welcome_user.html")
	fallbackPath := filepath.Join("..", "..", "..", "internal", "domain", "email", "templates", "welcome_user.html")

	templatePath := primaryPath
	_, err := os.Stat(primaryPath)

	if os.IsNotExist(err) {
		templatePath = fallbackPath
	}

	var tmpl *template.Template
	tmpl, err = template.ParseFiles(templatePath)

	if err != nil {
		panic(fmt.Sprintf("failed to parse welcome email template: %v", err))
	}

	return &SendWelcome{
		dialer:          mail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password),
		emailRepository: emailRepository,
		template:        tmpl,
	}
}

func (action *SendWelcome) Execute(to, name, verificationURL, eventID string) error {
	emailRecord := &model.Email{
		EventID: eventID,
		To:      to,
		Subject: "Verify your email - Go App",
		Status:  model.Pending,
		Type:    model.WelcomeEmail,
	}

	err := action.emailRepository.Create(emailRecord)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		}

		return err
	}

	var body bytes.Buffer
	dataWelcome := dtos.WelcomeEmail{
		Name:            name,
		Email:           to,
		VerificationURL: verificationURL,
	}

	err = action.template.Execute(&body, dataWelcome)

	if err != nil {
		_ = action.emailRepository.UpdateStatus(emailRecord.ID, model.Failed)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	message := mail.NewMessage()
	message.SetHeader("From", "no-reply@go-app.com")
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Verify your email - Go App")
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
