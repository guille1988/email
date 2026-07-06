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

type slowHandler struct {
	started  chan struct{}
	finished chan struct{}
}

func (handler *slowHandler) Handle(_ []byte, _ string) error {
	close(handler.started)
	time.Sleep(2 * time.Second)
	close(handler.finished)
	return nil
}

func TestKafkaConsumerShutdownDrainsInFlightWork(test *testing.T) {
	integration.TestCase(test, "it should let an in-flight handler finish before Close returns", func(test *testing.T) {
		brokers := integration.TestConfig.Kafka.Brokers
		topic := fmt.Sprintf("test.shutdown.drain.%d", time.Now().UnixNano())
		group := fmt.Sprintf("test-shutdown-group-%d", time.Now().UnixNano())

		producer, err := kgo.NewClient(kgo.SeedBrokers(brokers), kgo.AllowAutoTopicCreation())
		assert.NoError(test, err)
		defer producer.Close()

		produceCtx, produceCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer produceCancel()

		produceResults := producer.ProduceSync(produceCtx, &kgo.Record{Topic: topic, Value: []byte("payload")})
		assert.NoError(test, produceResults.FirstErr())

		handler := &slowHandler{started: make(chan struct{}), finished: make(chan struct{})}
		consumer := messaging.NewKafkaConsumer(brokers, 60000, 1)
		err = consumer.Register(group, "", "", topic, handler)
		assert.NoError(test, err)

		ctx, cancel := context.WithCancel(context.Background())
		err = consumer.StartAll(ctx)
		assert.NoError(test, err)

		select {
		case <-handler.started:
		case <-time.After(20 * time.Second):
			test.Fatal("timed out waiting for the handler to start")
		}

		// Simulate a shutdown signal arriving while the handler is still
		// mid-flight (asleep for 2s). Close must not return before that
		// elapses, and the handler must have already finished by then.
		closeStartedAt := time.Now()
		cancel()
		err = consumer.Close()
		closeElapsed := time.Since(closeStartedAt)
		assert.NoError(test, err)

		assert.GreaterOrEqual(test, closeElapsed, 2*time.Second, "Close() must block for at least as long as the in-flight handler takes")

		select {
		case <-handler.finished:
			// good: the handler had already completed by the time Close returned
		default:
			test.Fatal("Close() returned before the in-flight handler finished")
		}
	})
}
