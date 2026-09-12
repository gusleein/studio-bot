package domain

import (
	"time"

	"github.com/google/uuid"
)

// Rent - запись по аренде помещения
type Rent struct {
	ID       uuid.UUID
	ClientID uuid.UUID

	StartsAt time.Time // время начала
	EndsAt   time.Time // фактическое время завершения

	IsCancelled bool
	IsPaid      bool // произведена оплата

	PricePerHour int // цена за час
	PaidDuration int // количество предоплаченных часов

	BonusDuration int // дополнительное время сверх предоплаченного

	DoorCode string // код выдается на время сессии для кодового замка

	Type string // тип аренды "dj, production"

	PaidAt time.Time

	Notes string

	CreatedAt time.Time
	UpdatedAt time.Time
}
