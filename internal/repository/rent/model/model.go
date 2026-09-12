package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// RentModel — строка rents.
type RentModel struct {
	ID            uuid.UUID    `db:"id"`
	ClientID      uuid.UUID    `db:"client_id"`
	StartsAt      time.Time    `db:"starts_at"`
	EndsAt        sql.NullTime `db:"ends_at"`
	IsCancelled   bool         `db:"is_cancelled"`
	IsPaid        bool         `db:"is_paid"`
	PricePerHour  int          `db:"price_per_hour"`
	PaidDuration  int          `db:"paid_duration"`
	BonusDuration int          `db:"bonus_duration"`
	DoorCode      string       `db:"door_code"`
	Type          string       `db:"type"`
	PaidAt        sql.NullTime `db:"paid_at"`
	Notes         string       `db:"notes"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
}
