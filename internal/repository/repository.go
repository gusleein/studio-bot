package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// TelegramUserRepository — хранилище профилей Telegram.
type TelegramUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TelegramUser, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramUser, error)
	Create(ctx context.Context, user *domain.TelegramUser) (*domain.TelegramUser, error)
	Update(ctx context.Context, user *domain.TelegramUser) (*domain.TelegramUser, error)
}

// ClientRepository — хранилище клиентов аренды.
// Rents и Payments в агрегате не заполняются: их грузит сервис через RentRepository / PaymentRepository.
type ClientRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error)
	Create(ctx context.Context, client *domain.Client) (*domain.Client, error)
	Update(ctx context.Context, client *domain.Client) (*domain.Client, error)
}

// RentRepository — хранилище сессий аренды.
type RentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Rent, error)
	ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Rent, error)
	Create(ctx context.Context, rent *domain.Rent) (*domain.Rent, error)
	Update(ctx context.Context, rent *domain.Rent) (*domain.Rent, error)
}

// ProductRepository — каталог доп. товаров.
type ProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	List(ctx context.Context) ([]*domain.Product, error)
	Create(ctx context.Context, product *domain.Product) (*domain.Product, error)
}

// PaymentRepository — попытки оплаты клиента.
type PaymentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByTributeOrderUUID(ctx context.Context, orderUUID string) (*domain.Payment, error)
	ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Payment, error)
	Create(ctx context.Context, payment *domain.Payment) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) (*domain.Payment, error)
}
