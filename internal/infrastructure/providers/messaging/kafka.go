package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type MessageHandler interface {
	Handle(body []byte, eventID string) error
}

type topicEntry struct {
	topic   string
	handler MessageHandler
}

type KafkaConsumer struct {
	brokers          []string
	groupID          string
	entries          []topicEntry
	client           *kgo.Client
	rebalanceTimeout time.Duration
	workerPoolSize   int
	done             chan struct{}
}

func NewKafkaConsumer(brokers string, rebalanceTimeoutMs int, workerPoolSize int) *KafkaConsumer {
	return &KafkaConsumer{
		brokers:          []string{brokers},
		rebalanceTimeout: time.Duration(rebalanceTimeoutMs) * time.Millisecond,
		workerPoolSize:   workerPoolSize,
	}
}

func (consumer *KafkaConsumer) Register(queue, _, _, routingKey string, handler MessageHandler) error {
	if consumer.groupID == "" {
		consumer.groupID = queue
	}
	consumer.entries = append(consumer.entries, topicEntry{topic: routingKey, handler: handler})
	return nil
}

func commitWithRetry(ctx context.Context, cl *kgo.Client, label string) {
	for i := 0; i < 3; i++ {
		err := cl.CommitMarkedOffsets(ctx)

		if err != nil {
			slog.Error(label, "attempt", i+1, "error", err)
			time.Sleep(100 * time.Millisecond)

			continue
		}

		return
	}
}

func (consumer *KafkaConsumer) StartAll(ctx context.Context) error {
	topics := make([]string, 0, len(consumer.entries))
	for _, e := range consumer.entries {
		topics = append(topics, e.topic)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(consumer.brokers...),
		kgo.ConsumerGroup(consumer.groupID),
		kgo.ConsumeTopics(topics...),
		kgo.AllowAutoTopicCreation(),
		kgo.BlockRebalanceOnPoll(),
		kgo.RebalanceTimeout(consumer.rebalanceTimeout),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.AutoCommitMarks(),
		kgo.OnPartitionsRevoked(func(ctx context.Context, cl *kgo.Client, _ map[string][]int32) {
			commitWithRetry(ctx, cl, "failed to commit offsets on revoke")
		}),
	)
	if err != nil {
		return err
	}

	consumer.client = client
	consumer.done = make(chan struct{})

	handlers := make(map[string]MessageHandler, len(consumer.entries))

	for _, entries := range consumer.entries {
		handlers[entries.topic] = entries.handler
	}

	go func() {
		defer close(consumer.done)

		for {
			fetches := consumer.client.PollFetches(ctx)
			if ctx.Err() != nil {
				consumer.client.AllowRebalance()
				return
			}

			errs := fetches.Errors()

			if len(errs) > 0 {
				for _, theError := range errs {
					slog.Error("kafka fetch error", "error", theError.Err)
				}
				consumer.client.AllowRebalance()
				continue
			}

			sem := make(chan struct{}, consumer.workerPoolSize)
			var waitGroup sync.WaitGroup

			fetches.EachRecord(func(record *kgo.Record) {
				waitGroup.Add(1)
				sem <- struct{}{}
				go func(rec *kgo.Record) {
					defer waitGroup.Done()
					defer func() { <-sem }()

					slog.Info("message received from kafka", "topic", rec.Topic)

					eventID := fmt.Sprintf("%s:%d:%d", rec.Topic, rec.Partition, rec.Offset)
					handler, ok := handlers[rec.Topic]

					if ok {
						handlerErr := handler.Handle(rec.Value, eventID)

						if handlerErr != nil {
							slog.Error("handler error", "error", handlerErr)
						} else {
							consumer.client.MarkCommitRecords(rec)
						}
					}
				}(record)
			})
			waitGroup.Wait()

			commitWithRetry(ctx, consumer.client, "failed to commit offsets")
			consumer.client.AllowRebalance()
		}
	}()

	return nil
}

/*
Close waits for the in-flight batch (if any) to finish processing and
commit — signaled by the poll loop exiting after ctx is canceled — before
closing the underlying client. Callers must cancel StartAll's ctx before
calling Close, otherwise this blocks until that happens.
*/
func (consumer *KafkaConsumer) Close() error {
	if consumer.done != nil {
		<-consumer.done
	}

	if consumer.client != nil {
		consumer.client.Close()
	}

	return nil
}
