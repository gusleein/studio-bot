package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
)

type AppSettingsService interface {
	Get(ctx context.Context) (domain.AppSettings, error)
	UpdateAppName(ctx context.Context, appName string) error
	UpdateOpeningHours(ctx context.Context, openingHours domain.OpeningHours) error
	UpdateCustomPrices(ctx context.Context, customPrices []domain.CustomPrice) error
	UpdateBotAdmins(ctx context.Context, botAdmins []domain.BotAdmin) error
}

// ClientUpsert — данные Telegram при /start.
type ClientUpsert struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
}

// ClientService — регистрация клиента и Telegram-профиля.
type ClientService interface {
	GetOrCreate(ctx context.Context, in ClientUpsert) (domain.Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Client, error)
	SavePhone(ctx context.Context, telegramID int64, phone string) error
}

// RentService — бронирование и просмотр аренды студии.
type RentService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Rent, error)
	Upcoming(ctx context.Context, clientID uuid.UUID) (*domain.Rent, error)
	History(ctx context.Context, clientID uuid.UUID) ([]*domain.Rent, error)
	Occupied(ctx context.Context, from, to time.Time) ([]*domain.Rent, error)
	CreateUnpaid(ctx context.Context, rent *domain.Rent) (*domain.Rent, error)
	ConfirmPaid(ctx context.Context, id uuid.UUID) (*domain.Rent, error)
	CancelPaid(ctx context.Context, isPaid bool, id uuid.UUID) (*domain.Rent, error)
	CancelUnpaid(ctx context.Context, id uuid.UUID) error
}
