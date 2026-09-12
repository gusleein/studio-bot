package domain

import (
	"github.com/google/uuid"
	"time"
)

// Client - запись о клиенте по аренде студии
type Client struct {
	ID     uuid.UUID
	TgUser TelegramUser

	// заполняем через join
	Rents     []Rent
	Payments  []Payment
	LastVisit time.Time

	Notes     string // заметка о клиенте
	IsBlocked bool   // заблокирован доступ в помещение

	CreatedAt time.Time
	UpdatedAt time.Time
}
