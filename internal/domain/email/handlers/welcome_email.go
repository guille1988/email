package handlers

import (
	"email/internal/domain/email/actions"
	"email/internal/shared/messaging/rabbitmq/dtos"
	"encoding/json"
	"fmt"
)

type WelcomeEmail struct {
	action *actions.SendWelcome
}

func NewWelcomeEmail(action *actions.SendWelcome) *WelcomeEmail {
	return &WelcomeEmail{
		action: action,
	}
}

func (handler *WelcomeEmail) Handle(body []byte) error {
	var dto dtos.WelcomeEmail
	err := json.Unmarshal(body, &dto)

	if err != nil {
		return fmt.Errorf("failed to unmarshal welcome email dto: %w", err)
	}

	return handler.action.Execute(dto.Email, dto.Name)
}
