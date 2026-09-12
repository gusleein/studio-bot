package student

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// Service — реализация StudentService.
type Service struct {
	repo repository.StudentRepository
	log  *logger.Logger
}

// New создаёт новый сервис студентов.
func New(repo repository.StudentRepository, log *logger.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// GetOrCreate возвращает студента по telegram_id, создаёт нового если не найден.
func (s *Service) GetOrCreate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.Student, error) {
	student, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("поиск студента по telegram_id %d: %w", telegramID, err)
		}

		now := time.Now()
		newStudent := &domain.Student{
			ID:         uuid.New(),
			TelegramID: telegramID,
			FirstName:  firstName,
			LastName:   lastName,
			Username:   username,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		student, err = s.repo.Create(ctx, newStudent)
		if err != nil {
			return nil, fmt.Errorf("создание студента: %w", err)
		}

		s.log.Info("создан новый студент",
			zap.String("id", student.ID.String()),
			zap.Int64("telegram_id", telegramID),
		)
	}

	return student, nil
}

// GetByTelegramID возвращает студента по telegram_id.
func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Student, error) {
	student, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("получение студента по telegram_id %d: %w", telegramID, err)
	}
	return student, nil
}

// GetByID возвращает студента по UUID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("получение студента по id %s: %w", id, err)
	}
	return student, nil
}

// Update обновляет данные студента.
func (s *Service) Update(ctx context.Context, student *domain.Student) (*domain.Student, error) {
	updated, err := s.repo.Update(ctx, student)
	if err != nil {
		return nil, fmt.Errorf("обновление студента %s: %w", student.ID, err)
	}

	s.log.Info("студент обновлён", zap.String("id", student.ID.String()))
	return updated, nil
}
