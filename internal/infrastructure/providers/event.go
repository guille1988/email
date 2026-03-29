package providers

import (
	"email/internal/infrastructure/rabbitmq"
	"email/internal/shared/events"
)

type EventProvider struct {
	listeners map[string]rabbitmq.EventListener
	events    map[string]events.Event
}

func NewEventProvider() *EventProvider {
	return &EventProvider{
		listeners: make(map[string]rabbitmq.EventListener),
		events:    make(map[string]events.Event),
	}
}

func (provider *EventProvider) Register(event events.Event, listener rabbitmq.EventListener) {
	provider.listeners[event.RoutingKey()] = listener
	provider.events[event.RoutingKey()] = event
}

func (provider *EventProvider) GetListener(routingKey string) (rabbitmq.EventListener, bool) {
	l, ok := provider.listeners[routingKey]
	return l, ok
}

func (provider *EventProvider) GetEvent(routingKey string) (events.Event, bool) {
	event, ok := provider.events[routingKey]

	return event, ok
}

func (provider *EventProvider) GetRegisteredRoutingKeys() []string {
	keys := make([]string, 0, len(provider.listeners))

	for routingKey := range provider.listeners {
		keys = append(keys, routingKey)
	}

	return keys
}
