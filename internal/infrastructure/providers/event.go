package providers

import (
	"email/internal/domain/email/events"
	"email/internal/infrastructure/rabbitmq"
)

type EventProvider struct {
	listeners map[string]rabbitmq.EventListener
}

func NewEventProvider() *EventProvider {
	return &EventProvider{
		listeners: make(map[string]rabbitmq.EventListener),
	}
}

func (provider *EventProvider) Register(event events.Event, listener rabbitmq.EventListener) {
	provider.listeners[event.RoutingKey()] = listener
}

func (provider *EventProvider) GetListener(routingKey string) (rabbitmq.EventListener, bool) {
	l, ok := provider.listeners[routingKey]
	return l, ok
}
