package providers

import (
	"email/internal/infrastructure/rabbitmq"
	"email/internal/shared/events"
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

func (provider *EventProvider) GetRegisteredRoutingKeys() []string {
	keys := make([]string, 0, len(provider.listeners))

	for routingKey := range provider.listeners {
		keys = append(keys, routingKey)
	}

	return keys
}
