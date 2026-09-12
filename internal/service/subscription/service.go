package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/database"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// errAlreadyProcessed — платёж уже был обработан ранее (идемпотентность).
var errAlreadyProcessed = errors.New("платёж уже обработан")

// Service — реализация SubscriptionService.
type Service struct {
	db          *sqlx.DB
	subRepo     repository.SubscriptionRepository
	paymentRepo repository.PaymentRepository
	log         *logger.Logger
}

// New создаёт новый сервис подписок.
func New(db *sqlx.DB, subRepo repository.SubscriptionRepository, paymentRepo repository.PaymentRepository, log *logger.Logger) *Service {
	return &Service{
		db:          db,
		subRepo:     subRepo,
		paymentRepo: paymentRepo,
		log:         log,
	}
}

// GetActive возвращает все активные подписки студента.
func (s *Service) GetActive(ctx context.Context, studentID uuid.UUID) ([]*domain.Subscription, error) {
	subs, err := s.subRepo.GetActiveByStudentID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("получение активных подписок: %w", err)
	}
	return subs, nil
}

// IsSubscribed проверяет, есть ли у студента активная подписка заданного типа.
// Подписка bundle даёт доступ ко всем типам.
func (s *Service) IsSubscribed(ctx context.Context, studentID uuid.UUID, t domain.SubscriptionType) (bool, error) {
	subs, err := s.GetActive(ctx, studentID)
	if err != nil {
		return false, err
	}
	for _, sub := range subs {
		if sub.Type == t || sub.Type == domain.SubscriptionTypeBundle {
			return true, nil
		}
	}
	return false, nil
}

// Activate активирует (или продлевает) подписку через транзакцию.
// Если подписка уже активна — продлевает от текущей даты окончания.
func (s *Service) Activate(ctx context.Context, req service.ActivateRequest) (*domain.Subscription, error) {
	var result *domain.Subscription

	err := database.WithTx(ctx, s.db, func(txCtx context.Context) error {
		// Проверка идемпотентности внутри транзакции
		exists, err := s.paymentRepo.ExistsByTributeData(txCtx, req.TributeUserID, req.TributeProductID, req.Amount)
		if err != nil {
			return fmt.Errorf("проверка платежа: %w", err)
		}
		if exists {
			return errAlreadyProcessed
		}

		// Определяем базу для продления: если подписка активна — продлеваем от её expires_at
		now := time.Now()
		expiresAt := now.AddDate(0, 0, req.DurationDays)

		existing, err := s.subRepo.GetActiveByStudentID(txCtx, req.StudentID)
		if err == nil {
			for _, sub := range existing {
				if sub.Type == req.Type || req.Type == domain.SubscriptionTypeBundle {
					if sub.ExpiresAt.After(now) {
						expiresAt = sub.ExpiresAt.AddDate(0, 0, req.DurationDays)
					}
					break
				}
			}
		}

		// Создаём/обновляем подписку (upsert)
		sub := &domain.Subscription{
			ID:               uuid.New(),
			StudentID:        req.StudentID,
			Type:             req.Type,
			Status:           domain.SubscriptionStatusActive,
			TributeProductID: req.TributeProductID,
			StartedAt:        now,
			ExpiresAt:        expiresAt,
		}

		upserted, err := s.subRepo.Upsert(txCtx, sub)
		if err != nil {
			return fmt.Errorf("upsert подписки: %w", err)
		}

		// Создаём запись о платеже
		payment := &domain.Payment{
			ID:               uuid.New(),
			TributeUserID:    req.TributeUserID,
			TributeProductID: req.TributeProductID,
			Amount:           req.Amount,
		}

		if _, err = s.paymentRepo.Create(txCtx, payment); err != nil {
			return fmt.Errorf("создание платежа: %w", err)
		}

		result = upserted
		return nil
	})
	if err != nil {
		if errors.Is(err, errAlreadyProcessed) {
			s.log.Info("платёж уже обработан, пропускаем",
				zap.Int64("tribute_user_id", req.TributeUserID),
				zap.Int("tribute_product_id", req.TributeProductID),
			)
			// Возвращаем текущую подписку
			subs, _ := s.subRepo.GetActiveByStudentID(ctx, req.StudentID)
			for _, sub := range subs {
				if sub.Type == req.Type {
					return sub, nil
				}
			}
			return nil, nil
		}
		return nil, err
	}

	s.log.Info("подписка активирована",
		zap.String("student_id", req.StudentID.String()),
		zap.String("type", string(req.Type)),
		zap.Time("expires_at", result.ExpiresAt),
	)
	return result, nil
}

// GetActiveSubscribersTelegramIDs возвращает telegram_id всех активных подписчиков заданного типа.
func (s *Service) GetActiveSubscribersTelegramIDs(ctx context.Context, t domain.SubscriptionType) ([]int64, error) {
	ids, err := s.subRepo.GetActiveSubscribersTelegramIDs(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("получение telegram_id подписчиков: %w", err)
	}
	return ids, nil
}

// StartExpiryJob запускает фоновую горутину, которая раз в час помечает истёкшие подписки.
func (s *Service) StartExpiryJob(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		s.log.Info("запуск задачи истечения подписок")
		for {
			select {
			case <-ctx.Done():
				s.log.Info("задача истечения подписок остановлена")
				return
			case <-ticker.C:
				if err := s.subRepo.ExpireOld(ctx); err != nil {
					s.log.Error("ошибка при истечении подписок", zap.Error(err))
				} else {
					s.log.Debug("задача истечения подписок выполнена")
				}
			}
		}
	}()
}
