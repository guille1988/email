package model

import (
	"email/internal/infrastructure/config"
	"time"
)

type EmailStatus string

const (
	Pending EmailStatus = "pending"
	Sent    EmailStatus = "sent"
	Failed  EmailStatus = "failed"
)

type Email struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	To        string      `gorm:"size:255;index;not null" json:"to"`
	Subject   string      `gorm:"size:255;not null" json:"subject"`
	Body      string      `gorm:"type:text;not null" json:"body"`
	Status    EmailStatus `gorm:"size:50;default:'pending';index" json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// TableName especifica el nombre de la tabla para el modelo Email
func (Email) TableName() string {
	return "emails"
}

// ConnectionName especifica la conexión de base de datos para el modelo Email
func (Email) ConnectionName() config.ConnectionName {
	return config.Default
}
