package handlers

import (
	"email/internal/domain/email/actions"
	"encoding/json"
	"fmt"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"
)

type StressEmail struct {
	action *actions.SendStress
}

func NewStressEmail(action *actions.SendStress) *StressEmail {
	return &StressEmail{
		action: action,
	}
}

func (handler *StressEmail) Handle(body []byte, eventID string) error {
	var dto dtos.StressEmail
	err := json.Unmarshal(body, &dto)

	if err != nil {
		return fmt.Errorf("failed to unmarshal stress email dto: %w", err)
	}

	return handler.action.Execute(dto.Email, dto.Name, eventID)
}
