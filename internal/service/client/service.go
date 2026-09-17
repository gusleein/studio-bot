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

func (s *Service) GetOrCreate(ctx context.Context, in service.ClientUpsert) (result domain.Client, err error) {
	user, err := s.users.GetByTelegramID(ctx, in.TelegramID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		err = fmt.Errorf("failed to get user by telegram id: %w", err)
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		err = database.WithTx(ctx, s.db, func(ctx context.Context) (err1 error) {
			created, err1 := s.users.Create(ctx, domain.TelegramUser{
				TelegramId: in.TelegramID,
				Username:   in.Username,
				FirstName:  in.FirstName,
				LastName:   in.LastName,
			})
			if err1 != nil {
				return fmt.Errorf("создание telegram user: %w", err1)
			}

			user = created

			_, err1 = s.clients.Create(ctx, domain.Client{TgUser: created})
			if err1 != nil {
				return fmt.Errorf("создание клиента: %w", err1)
			}
			return
		})
		if err != nil {
			return
		}

		s.log.Info("создан клиент",
			zap.String("user_id", user.Id.String()),
			zap.Int64("telegram_id", in.TelegramID),
		)
	} else {
		user.Username = in.Username
		user.FirstName = in.FirstName
		user.LastName = in.LastName

		if _, err = s.users.Update(ctx, user); err != nil {
			err = fmt.Errorf("обновление telegram user: %w", err)
			return
		}
	}

	client, err := s.clients.GetByTelegramID(ctx, in.TelegramID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			err = fmt.Errorf("failed to get client by telegram id: %w", err)
			return
		}

		client, err = s.clients.Create(ctx, domain.Client{TgUser: user})
		if err != nil {
			err = fmt.Errorf("failed to create client: %w", err)
			return
		}
	}

	result = client
	return
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (result domain.Client, err error) {
	result, err = s.clients.GetByID(ctx, id)
	return
}

func (s *Service) SavePhone(ctx context.Context, telegramID int64, phone string) (err error) {
	user, err := s.users.GetByTelegramID(ctx, telegramID)
	if err != nil {
		err = fmt.Errorf("failed to get user by telegram id: %w", err)
		return
	}

	user.Phone = phone
	if _, err := s.users.Update(ctx, user); err != nil {
		err = fmt.Errorf("failed to update user: %w", err)
	}
	return
}
