package emails

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/model"
	"fmt"
	"testing"
	"time"

	"email/tests/integration"

	"github.com/stretchr/testify/assert"
)

func TestEmailDedupRetry(test *testing.T) {
	integration.TestCase(test, "it should retry sending when a prior attempt was left pending, instead of dropping the email", func(test *testing.T) {
		to := fmt.Sprintf("retry-pending-%d@example.com", time.Now().UnixNano())
		name := "Retry Pending"
		eventID := fmt.Sprintf("retry-pending-event-%d", time.Now().UnixNano())

		emailRepository := model.NewRepository(integration.TestApp.Container.DefaultConnection)

		staleAttempt := &model.Email{
			EventID: eventID,
			To:      to,
			Subject: "Verify your email - Go App",
			Status:  model.Pending,
			Type:    model.WelcomeEmail,
		}
		assert.NoError(test, emailRepository.Create(staleAttempt))

		sendWelcomeAction := actions.NewSendWelcome(integration.TestConfig.Mail, emailRepository)
		err := sendWelcomeAction.Execute(to, name, "http://localhost:3000/verify-email?token=test", eventID)
		assert.NoError(test, err)

		var emailRecord *model.Email
		emailRecord, err = emailRepository.FindByEventID(eventID)

		assert.NoError(test, err)
		assert.Equal(test, staleAttempt.ID, emailRecord.ID, "the same row must be reused, not a new one")
		assert.Equal(test, model.Sent, emailRecord.Status, "the stale pending row must end up sent, not silently dropped")
	})

	integration.TestCase(test, "it should not resend when the prior attempt already succeeded", func(test *testing.T) {
		to := fmt.Sprintf("retry-sent-%d@example.com", time.Now().UnixNano())
		name := "Retry Sent"
		eventID := fmt.Sprintf("retry-sent-event-%d", time.Now().UnixNano())

		emailRepository := model.NewRepository(integration.TestApp.Container.DefaultConnection)

		alreadySent := &model.Email{
			EventID: eventID,
			To:      to,
			Subject: "Verify your email - Go App",
			Body:    "already sent",
			Status:  model.Sent,
			Type:    model.WelcomeEmail,
		}
		assert.NoError(test, emailRepository.Create(alreadySent))

		sendWelcomeAction := actions.NewSendWelcome(integration.TestConfig.Mail, emailRepository)
		err := sendWelcomeAction.Execute(to, name, "http://localhost:3000/verify-email?token=test", eventID)
		assert.NoError(test, err)

		var emailRecord *model.Email
		emailRecord, err = emailRepository.FindByEventID(eventID)

		assert.NoError(test, err)
		assert.Equal(test, "already sent", emailRecord.Body, "an already-sent email must not be touched again")
	})
}
