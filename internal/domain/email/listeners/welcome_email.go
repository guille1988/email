package listeners

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WelcomeEmail struct {
	Action *actions.SendWelcome
}

func (listener *WelcomeEmail) Handle(delivery amqp.Delivery) error {
	var event events.UserRegistered
	err := event.FromJson(delivery.Body)

	if err != nil {
		return err
	}

	return listener.Action.Execute(event.Email, event.Name)
}
