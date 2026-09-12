package service

import (
	"context"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// ClientUpsert — данные Telegram при /start.
type ClientUpsert struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
}

// ClientService — регистрация клиента и Telegram-профиля.
type ClientService interface {
	GetOrCreate(ctx context.Context, in ClientUpsert) (*domain.Client, error)
	SavePhone(ctx context.Context, telegramID int64, phone string) error
}
