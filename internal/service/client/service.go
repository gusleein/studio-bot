package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/database"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

type Service struct {
	db      *sqlx.DB
	users   repository.TelegramUserRepository
	clients repository.ClientRepository
	log     *logger.Logger
}

var _ service.ClientService = (*Service)(nil)

func New(
	db *sqlx.DB,
	users repository.TelegramUserRepository,
	clients repository.ClientRepository,
	log *logger.Logger,
) *Service {
	return &Service{db: db, users: users, clients: clients, log: log}
}

func (s *Service) GetOrCreate(ctx context.Context, in service.ClientUpsert) (*domain.Client, error) {
	user, err := s.users.GetByTelegramID(ctx, in.TelegramID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("поиск telegram user: %w", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		if err := database.WithTx(ctx, s.db, func(ctx context.Context) error {
			created, err := s.users.Create(ctx, &domain.TelegramUser{
				TelegramId: in.TelegramID,
				Username:   in.Username,
				FirstName:  in.FirstName,
				LastName:   in.LastName,
			})
			if err != nil {
				return fmt.Errorf("создание telegram user: %w", err)
			}
			user = created

			_, err = s.clients.Create(ctx, &domain.Client{TgUser: *created})
			if err != nil {
				return fmt.Errorf("создание клиента: %w", err)
			}
			return nil
		}); err != nil {
			return nil, err
		}

		s.log.Info("создан клиент",
			zap.String("user_id", user.Id.String()),
			zap.Int64("telegram_id", in.TelegramID),
		)
	} else {
		user.Username = in.Username
		user.FirstName = in.FirstName
		user.LastName = in.LastName
		if _, err := s.users.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("обновление telegram user: %w", err)
		}
	}

	client, err := s.clients.GetByTelegramID(ctx, in.TelegramID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("получение клиента: %w", err)
		}
		client, err = s.clients.Create(ctx, &domain.Client{TgUser: *user})
		if err != nil {
			return nil, fmt.Errorf("создание клиента для существующего user: %w", err)
		}
	}
	return client, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	client, err := s.clients.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("получение клиента: %w", err)
	}
	return client, nil
}

func (s *Service) SavePhone(ctx context.Context, telegramID int64, phone string) error {
	user, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("поиск telegram user: %w", err)
	}

	user.Phone = phone
	if _, err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("сохранение телефона: %w", err)
	}
	return nil
}
