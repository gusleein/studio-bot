package model

import (
	"time"

	"github.com/google/uuid"
)

// TelegramUserModel — строка telegram_users.
type TelegramUserModel struct {
	ID         uuid.UUID `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Phone      string    `db:"phone"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
