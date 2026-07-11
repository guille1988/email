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

func TestStressEmailModule(test *testing.T) {
	integration.TestCase(test, "it should send a stress email when receiving a message", func(test *testing.T) {
		to := fmt.Sprintf("stress-%d@example.com", time.Now().UnixNano())
		name := "Stress Tester"

		emailRepository := model.NewRepository(integration.TestApp.Container.DefaultConnection)
		sendStressAction := actions.NewSendStress(integration.TestConfig.Mail, emailRepository)
		handler := handlers.NewStressEmail(sendStressAction)
		body, _ := json.Marshal(dtos.StressEmail{Email: to, Name: name})
		err := handler.Handle(body, "0:0")
		assert.NoError(test, err)

		var emailRecord *model.Email
		emailRecord, err = emailRepository.FindByTo(to)

		assert.NoError(test, err)
		assert.Equal(test, model.Sent, emailRecord.Status)
		assert.Equal(test, "Stress Test — Go App", emailRecord.Subject)
		assert.Equal(test, model.StressEmail, emailRecord.Type)

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
		assert.True(test, found, "stress email not found in Mailpit for recipient %s", to)
	})
}
