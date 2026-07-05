package emails

import (
	"context"
	"email/internal/infrastructure/providers/messaging"
	"email/tests/integration"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
)

type failingHandler struct {
	calls chan string
}

func (handler *failingHandler) Handle(_ []byte, eventID string) error {
	select {
	case handler.calls <- eventID:
	default:
	}

	return fmt.Errorf("simulated handler failure")
}

type successHandler struct {
	calls chan string
}

func (handler *successHandler) Handle(_ []byte, eventID string) error {
	select {
	case handler.calls <- eventID:
	default:
	}

	return nil
}

func TestKafkaCommitMarks(test *testing.T) {
	integration.TestCase(test, "it should not commit the offset for a message whose handler fails, so a fresh consumer in the same group receives it again", func(test *testing.T) {
		brokers := integration.TestConfig.Kafka.Brokers
		topic := fmt.Sprintf("test.commit.marks.%d", time.Now().UnixNano())
		group := fmt.Sprintf("test-commit-group-%d", time.Now().UnixNano())

		producer, err := kgo.NewClient(kgo.SeedBrokers(brokers), kgo.AllowAutoTopicCreation())
		assert.NoError(test, err)
		defer producer.Close()

		produceCtx, produceCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer produceCancel()

		produceResults := producer.ProduceSync(produceCtx, &kgo.Record{Topic: topic, Value: []byte("payload")})
		assert.NoError(test, produceResults.FirstErr())

		firstHandler := &failingHandler{calls: make(chan string, 10)}
		consumerOne := messaging.NewKafkaConsumer(brokers, 60000, 1)
		err = consumerOne.Register(group, "", "", topic, firstHandler)
		assert.NoError(test, err)

		ctxOne, cancelOne := context.WithCancel(context.Background())
		err = consumerOne.StartAll(ctxOne)
		assert.NoError(test, err)

		var firstEventID string
		select {
		case firstEventID = <-firstHandler.calls:
		case <-time.After(20 * time.Second):
			test.Fatal("timed out waiting for the first (failing) handler invocation")
		}

		cancelOne()
		err = consumerOne.Close()
		assert.NoError(test, err)

		secondHandler := &successHandler{calls: make(chan string, 10)}
		consumerTwo := messaging.NewKafkaConsumer(brokers, 60000, 1)
		err = consumerTwo.Register(group, "", "", topic, secondHandler)
		assert.NoError(test, err)

		ctxTwo, cancelTwo := context.WithCancel(context.Background())
		defer cancelTwo()
		err = consumerTwo.StartAll(ctxTwo)
		assert.NoError(test, err)

		select {
		case redeliveredEventID := <-secondHandler.calls:
			assert.Equal(test, firstEventID, redeliveredEventID, "expected the same message to be redelivered")
		case <-time.After(20 * time.Second):
			test.Fatal("message was not redelivered to a fresh consumer in the same group — offset was committed despite the handler failing")
		}

		_ = consumerTwo.Close()
	})
}
