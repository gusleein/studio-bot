package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// StudentService — интерфейс сервиса студентов.
type StudentService interface {
	GetOrCreate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.Student, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Student, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error)
	Update(ctx context.Context, student *domain.Student) (*domain.Student, error)
}

// ScheduleService — интерфейс сервиса расписания.
type ScheduleService interface {
	GetWeek(ctx context.Context, weekOffset int) ([]*domain.Schedule, error)
	GetMyWeek(ctx context.Context, studentID uuid.UUID, weekOffset int) ([]*domain.Schedule, error)
	List(ctx context.Context, from, to time.Time) ([]*domain.Schedule, error)
	ListByStudentID(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]*domain.Schedule, error)
}

// SubscriptionService — интерфейс сервиса подписок.
type SubscriptionService interface {
	GetActive(ctx context.Context, studentID uuid.UUID) ([]*domain.Subscription, error)
	IsSubscribed(ctx context.Context, studentID uuid.UUID, t domain.SubscriptionType) (bool, error)
	Activate(ctx context.Context, req ActivateRequest) (*domain.Subscription, error)
	GetActiveSubscribersTelegramIDs(ctx context.Context, t domain.SubscriptionType) ([]int64, error)
	StartExpiryJob(ctx context.Context)
}

// PackService — интерфейс сервиса паков.
type PackService interface {
	List(ctx context.Context, packType *domain.PackType) ([]*domain.Pack, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pack, error)
	Create(ctx context.Context, pack *domain.Pack) (*domain.Pack, error)
	SendToStudent(ctx context.Context, packID uuid.UUID, telegramID int64) error
	ListUnnotified(ctx context.Context) ([]*domain.Pack, error)
	MarkNotified(ctx context.Context, id uuid.UUID) error
}

// ActivateRequest — запрос на активацию подписки (используется в webhook и сервисе).
type ActivateRequest struct {
	StudentID        uuid.UUID
	Type             domain.SubscriptionType
	TributeProductID int
	TributeUserID    int64
	Amount           int
	Currency         string
	DurationDays     int
}
