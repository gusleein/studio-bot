package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ClientModel — строка clients для записи.
type ClientModel struct {
	ID             uuid.UUID    `db:"id"`
	TelegramUserID uuid.UUID    `db:"telegram_user_id"`
	Notes          string       `db:"notes"`
	IsBlocked      bool         `db:"is_blocked"`
	LastVisit      sql.NullTime `db:"last_visit"`
	CreatedAt      time.Time    `db:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at"`
}

// ClientWithUserRow — клиент + telegram_users (JOIN).
type ClientWithUserRow struct {
	ID             uuid.UUID    `db:"id"`
	TelegramUserID uuid.UUID    `db:"telegram_user_id"`
	Notes          string       `db:"notes"`
	IsBlocked      bool         `db:"is_blocked"`
	LastVisit      sql.NullTime `db:"last_visit"`
	CreatedAt      time.Time    `db:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at"`

	UserID        uuid.UUID `db:"user_id"`
	TelegramID    int64     `db:"telegram_id"`
	Username      string    `db:"username"`
	FirstName     string    `db:"first_name"`
	LastName      string    `db:"last_name"`
	Phone         string    `db:"phone"`
	UserCreatedAt time.Time `db:"user_created_at"`
	UserUpdatedAt time.Time `db:"user_updated_at"`
}
