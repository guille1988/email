package emails

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/handlers"
	"email/internal/domain/email/model"
	"email/internal/shared/messaging/rabbitmq/dtos"
	"email/tests/integration"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmailModule(test *testing.T) {
	integration.TestCase(test, "it should send a welcome email when receiving a message", func(test *testing.T) {
		to := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
		name := "Test User"

		emailRepo := model.NewRepository(integration.TestApp.Container.DefaultConnection)
		sendWelcomeAction := actions.NewSendWelcome(integration.TestConfig.Mail, emailRepo)
		handler := handlers.NewWelcomeEmail(sendWelcomeAction)
		body, _ := json.Marshal(dtos.WelcomeEmail{Email: to, Name: name, VerificationURL: "http://localhost:3000/verify-email?token=test"})
		err := handler.Handle(body)
		assert.NoError(test, err)

		emailRecord, err := emailRepo.FindByTo(to)
		assert.NoError(test, err)
		assert.Equal(test, model.Sent, emailRecord.Status)
		assert.Equal(test, "Verify your email - Go App", emailRecord.Subject)

		var resp *http.Response
		resp, err = http.Get(fmt.Sprintf("http://%s:%d/api/v1/messages",
			integration.TestConfig.Mail.Host, integration.MailpitApiPort))
		assert.NoError(test, err)

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)

		mailBody, _ := io.ReadAll(resp.Body)
		var mailpitResp struct {
			Messages []struct {
				To []struct {
					Address string `json:"Address"`
				} `json:"To"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(mailBody, &mailpitResp)

		found := false
		for _, msg := range mailpitResp.Messages {
			for _, recipient := range msg.To {
				if recipient.Address == to {
					found = true
					break
				}
			}
		}
		assert.True(test, found, "email not found in Mailpit for recipient %s", to)
	})
}
