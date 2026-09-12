package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// Service — реализация ScheduleService.
type Service struct {
	repo repository.ScheduleRepository
	log  *logger.Logger
}

// New создаёт новый сервис расписания.
func New(repo repository.ScheduleRepository, log *logger.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// weekBounds вычисляет понедельник 00:00 и воскресенье 23:59:59 для недели с заданным смещением.
// weekOffset=0 — текущая неделя, 1 — следующая, -1 — прошлая.
func weekBounds(weekOffset int) (time.Time, time.Time) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday-1)+weekOffset*7)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())
	return monday, sunday
}

// GetWeek возвращает расписание на заданную неделю (weekOffset=0 — текущая).
func (s *Service) GetWeek(ctx context.Context, weekOffset int) ([]*domain.Schedule, error) {
	from, to := weekBounds(weekOffset)
	schedules, err := s.repo.List(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("получение расписания на неделю (offset=%d): %w", weekOffset, err)
	}
	return schedules, nil
}

// GetMyWeek возвращает расписание конкретного студента на заданную неделю.
func (s *Service) GetMyWeek(ctx context.Context, studentID uuid.UUID, weekOffset int) ([]*domain.Schedule, error) {
	from, to := weekBounds(weekOffset)
	schedules, err := s.repo.ListByStudentID(ctx, studentID, from, to)
	if err != nil {
		return nil, fmt.Errorf("получение расписания студента на неделю (offset=%d): %w", weekOffset, err)
	}
	return schedules, nil
}

// List возвращает расписание в заданном диапазоне дат.
func (s *Service) List(ctx context.Context, from, to time.Time) ([]*domain.Schedule, error) {
	schedules, err := s.repo.List(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("получение расписания: %w", err)
	}
	return schedules, nil
}

// ListByStudentID возвращает расписание студента в заданном диапазоне дат.
func (s *Service) ListByStudentID(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]*domain.Schedule, error) {
	schedules, err := s.repo.ListByStudentID(ctx, studentID, from, to)
	if err != nil {
		return nil, fmt.Errorf("получение расписания студента: %w", err)
	}
	return schedules, nil
}
