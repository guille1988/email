package messaging

import (
	"email/internal/infrastructure/providers/messaging"
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopHandler struct{}

func (noopHandler) Handle(_ []byte, _ string) error { return nil }

func TestRegisterRejectsMismatchedGroupID(test *testing.T) {
	consumer := messaging.NewKafkaConsumer("kafka:9092", 60000, 1)

	assert.NoError(test, consumer.Register("service.a", "", "", "topic.one", noopHandler{}))
	assert.NoError(test, consumer.Register("service.a", "", "", "topic.two", noopHandler{}), "same group id on a second topic must be accepted")

	err := consumer.Register("service.b", "", "", "topic.three", noopHandler{})

	if assert.Error(test, err, "a different group id must be rejected instead of silently ignored") {
		assert.Contains(test, err.Error(), "mismatched group id")
	}

	/*
	 The rejected registration must not corrupt the consumer: the original
	 group id keeps working afterward.
	*/
	assert.NoError(test, consumer.Register("service.a", "", "", "topic.three", noopHandler{}))
}
