package emails

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/handlers"
	"email/internal/domain/email/model"
	"email/tests/integration"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"

	"github.com/stretchr/testify/assert"
)

func TestEmailModule(test *testing.T) {
	integration.TestCase(test, "it should send a welcome email when receiving a message", func(test *testing.T) {
		to := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
		name := "Test User"

		emailRepository := model.NewRepository(integration.TestApp.Container.DefaultConnection)
		sendWelcomeAction := actions.NewSendWelcome(integration.TestConfig.Mail, emailRepository)
		handler := handlers.NewWelcomeEmail(sendWelcomeAction)
		body, _ := json.Marshal(dtos.WelcomeEmail{Email: to, Name: name, VerificationURL: "http://localhost:3000/verify-email?token=test"})
		err := handler.Handle(body, "0:0")
		assert.NoError(test, err)

		var emailRecord *model.Email
		emailRecord, err = emailRepository.FindByTo(to)
		assert.NoError(test, err)
		assert.Equal(test, model.Sent, emailRecord.Status)
		assert.Equal(test, "Verify your email - Go App", emailRecord.Subject)
		assert.Equal(test, model.WelcomeEmail, emailRecord.Type)

		var response *http.Response
		response, err = http.Get(fmt.Sprintf("http://%s:%d/api/v1/messages",
			integration.TestConfig.Mail.Host, integration.MailpitApiPort))
		assert.NoError(test, err)

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				panic(err)
			}
		}(response.Body)

		mailBody, _ := io.ReadAll(response.Body)
		var mailpitResponse struct {
			Messages []struct {
				To []struct {
					Address string `json:"Address"`
				} `json:"To"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(mailBody, &mailpitResponse)

		found := false
		for _, msg := range mailpitResponse.Messages {
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
