package rent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

type Service struct {
	rents repository.RentRepository
	log   *logger.Logger
}

var _ service.RentService = (*Service)(nil)

func New(rents repository.RentRepository, log *logger.Logger) *Service {
	return &Service{rents: rents, log: log}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rent, error) {
	return s.rents.GetByID(ctx, id)
}

func (s *Service) Upcoming(ctx context.Context, clientID uuid.UUID) (*domain.Rent, error) {
	rent, err := s.rents.GetUpcomingByClientID(ctx, clientID, time.Now())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ближайшая аренда: %w", err)
	}
	return rent, nil
}

func (s *Service) History(ctx context.Context, clientID uuid.UUID) ([]*domain.Rent, error) {
	items, err := s.rents.ListByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("история аренды: %w", err)
	}
	return items, nil
}

func (s *Service) Occupied(ctx context.Context, from, to time.Time) ([]*domain.Rent, error) {
	items, err := s.rents.ListActiveInRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("занятые слоты: %w", err)
	}
	return items, nil
}

func (s *Service) CreateUnpaid(ctx context.Context, rent *domain.Rent) (*domain.Rent, error) {
	busy, err := s.rents.ListActiveInRange(ctx, rent.StartsAt, rent.EndsAt)
	if err != nil {
		return nil, fmt.Errorf("проверка слота: %w", err)
	}
	if len(busy) > 0 {
		return nil, domain.ErrSlotBusy
	}

	created, err := s.rents.Create(ctx, rent)
	if err != nil {
		return nil, fmt.Errorf("создание аренды: %w", err)
	}
	return created, nil
}

func (s *Service) ConfirmPaid(ctx context.Context, id uuid.UUID) (*domain.Rent, error) {
	rent, err := s.rents.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("аренда для подтверждения: %w", err)
	}
	if rent.IsCancelled {
		return nil, domain.ErrNotFound
	}
	if rent.IsPaid {
		return rent, nil
	}

	rent.IsPaid = true
	rent.PaidAt = time.Now()
	updated, err := s.rents.Update(ctx, rent)
	if err != nil {
		return nil, fmt.Errorf("подтверждение оплаты: %w", err)
	}
	return updated, nil
}

func (s *Service) CancelPaid(ctx context.Context, isPaid bool, id uuid.UUID) (*domain.Rent, error) {
	rent, err := s.rents.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("аренда для отмены: %w", err)
	}
	if rent.IsCancelled {
		return nil, domain.ErrNotFound
	}
	if rent.IsPaid {
		return rent, nil
	}

	// отменяем аренду
	rent.IsCancelled = true
	rent.IsPaid = isPaid
	if isPaid {
		rent.PaidAt = time.Now()
	}
	updated, err := s.rents.Update(ctx, rent)
	if err != nil {
		return nil, fmt.Errorf("подтверждение оплаты: %w", err)
	}
	return updated, nil
}

func (s *Service) CancelUnpaid(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	rent, err := s.rents.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("аренда для отмены: %w", err)
	}
	if rent.IsPaid {
		return nil
	}
	rent.IsCancelled = true
	if _, err := s.rents.Update(ctx, rent); err != nil {
		return fmt.Errorf("отмена аренды: %w", err)
	}
	return nil
}
