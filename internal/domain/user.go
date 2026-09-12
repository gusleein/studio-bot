package domain

import (
	"github.com/google/uuid"
	"time"
)

// TelegramUser - модель пользователя телеграмма для хранения состояиния и данных
type TelegramUser struct {
	Id         uuid.UUID
	TelegramId int64
	Username   string
	FirstName  string
	LastName   string
	Phone      string

	CreatedAt time.Time
	UpdatedAt time.Time
}
