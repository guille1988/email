package emails

import (
	"email/internal/domain/email/model"
	"email/tests/integration"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDuplicateEntry(test *testing.T) {
	integration.TestCase(test, "it should recognize a real MySQL duplicate-key error and reject unrelated errors", func(test *testing.T) {
		emailRepository := model.NewRepository(integration.TestApp.Container.DefaultConnection)

		first := &model.Email{EventID: "dup-check-event-id", To: "a@example.com", Subject: "s", Status: model.Pending, Type: model.WelcomeEmail}
		assert.NoError(test, emailRepository.Create(first))

		second := &model.Email{EventID: "dup-check-event-id", To: "b@example.com", Subject: "s", Status: model.Pending, Type: model.WelcomeEmail}
		err := emailRepository.Create(second)

		assert.Error(test, err)
		assert.True(test, model.IsDuplicateEntry(err), "expected a real MySQL duplicate-key error to be recognized")

		assert.False(test, model.IsDuplicateEntry(errors.New("some unrelated error")))
		assert.False(test, model.IsDuplicateEntry(nil))
	})
}
