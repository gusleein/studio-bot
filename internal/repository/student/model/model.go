package model

import (
	"time"

	"github.com/google/uuid"
)

// StudentModel — модель ученика для хранения в БД.
type StudentModel struct {
	ID         uuid.UUID `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Phone      string    `db:"phone"`
	Email      string    `db:"email"`
	IsActive   bool      `db:"is_active"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
