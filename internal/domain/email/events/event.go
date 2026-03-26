package events

type Event interface {
	RoutingKey() string
	FromJson(data []byte) error
}
