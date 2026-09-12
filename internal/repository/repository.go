package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// StudentRepository — интерфейс репозитория студентов.
type StudentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Student, error)
	Create(ctx context.Context, student *domain.Student) (*domain.Student, error)
	Update(ctx context.Context, student *domain.Student) (*domain.Student, error)
}

// ScheduleRepository — интерфейс репозитория расписания.
type ScheduleRepository interface {
	List(ctx context.Context, from, to time.Time) ([]*domain.Schedule, error)
	ListByStudentID(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]*domain.Schedule, error)
}

// SubscriptionRepository — интерфейс репозитория подписок.
type SubscriptionRepository interface {
	GetActiveByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.Subscription, error)
	GetActiveSubscribersTelegramIDs(ctx context.Context, t domain.SubscriptionType) ([]int64, error)
	Upsert(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error)
	ExpireOld(ctx context.Context) error
}

// PackRepository — интерфейс репозитория паков.
type PackRepository interface {
	List(ctx context.Context, packType *domain.PackType) ([]*domain.Pack, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pack, error)
	Create(ctx context.Context, pack *domain.Pack) (*domain.Pack, error)
	UpdateTelegramFileID(ctx context.Context, id uuid.UUID, fileID string) error
	ListUnnotified(ctx context.Context) ([]*domain.Pack, error)
	MarkNotified(ctx context.Context, id uuid.UUID) error
}

// PaymentRepository — интерфейс репозитория платежей.
type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) (*domain.Payment, error)
	ExistsByTributeData(ctx context.Context, tributeUserID int64, tributeProductID int, amount int) (bool, error)
}

type TeacherRepository interface {
	Create(ctx context.Context, t *domain.Teacher) (*domain.Teacher, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Teacher, error)
	GetAll(ctx context.Context) ([]*domain.Teacher, error)
}
