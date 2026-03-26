package events

import "encoding/json"

type UserRegistered struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (event *UserRegistered) RoutingKey() string {
	return "user.registered"
}

func (event *UserRegistered) FromJson(data []byte) error {
	return json.Unmarshal(data, event)
}
