package subscription

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/subscription/converter"
	"github.com/yourstudio/studio-bot/internal/repository/subscription/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	// createSubscriptionQuery — вставка новой подписки.
	createSubscriptionQuery = `
		INSERT INTO subscriptions
			(id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at)
		VALUES
			(:id, :student_id, :type, :status, :tribute_product_id, :started_at, :expires_at, :created_at, :updated_at)`

	// upsertSubscriptionQuery — вставка с обновлением при конфликте по (student_id, type).
	upsertSubscriptionQuery = `
		INSERT INTO subscriptions
			(id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at)
		VALUES
			(:id, :student_id, :type, :status, :tribute_product_id, :started_at, :expires_at, :created_at, :updated_at)
		ON CONFLICT (student_id, type) DO UPDATE
		SET
			status     = 'active',
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()`

	// getActiveByStudentIDQuery — активные подписки студента.
	getActiveByStudentIDQuery = `
		SELECT id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at
		FROM subscriptions
		WHERE student_id = $1
		  AND status = 'active'
		  AND expires_at > NOW()`

	// getByStudentAndTypeQuery — последняя подписка студента заданного типа.
	getByStudentAndTypeQuery = `
		SELECT id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at
		FROM subscriptions
		WHERE student_id = $1
		  AND type = $2
		LIMIT 1`

	// expireOldQuery — переводит просроченные активные подписки в статус expired.
	expireOldQuery = `
		UPDATE subscriptions
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'active'
		  AND expires_at <= NOW()`

	// getActiveSubscribersTelegramIDsQuery — telegram_id студентов с активной подпиской заданного типа.
	// Bundle-подписка покрывает любой тип.
	getActiveSubscribersTelegramIDsQuery = `
		SELECT DISTINCT s.telegram_id
		FROM subscriptions sub
		JOIN students s ON s.id = sub.student_id
		WHERE sub.status = 'active'
		  AND sub.expires_at > NOW()
		  AND (sub.type = $1 OR sub.type = 'bundle')
		  AND s.is_active = TRUE`
)

// SubscriptionRepo — репозиторий для работы с подписками.
type SubscriptionRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр SubscriptionRepo.
func New(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// Create создаёт новую подписку в БД.
func (r *SubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	db := database.GetDB(ctx, r.db)

	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	m := converter.ToModel(sub)

	if _, err := db.NamedExecContext(ctx, createSubscriptionQuery, m); err != nil {
		return nil, errors.Wrap(err, "create subscription")
	}

	return sub, nil
}

// Upsert вставляет подписку или обновляет существующую по (student_id, type).
// После upsert возвращает актуальное состояние из БД.
func (r *SubscriptionRepo) Upsert(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	db := database.GetDB(ctx, r.db)

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	m := converter.ToModel(sub)

	if _, err := db.NamedExecContext(ctx, upsertSubscriptionQuery, m); err != nil {
		return nil, errors.Wrap(err, "upsert subscription")
	}

	// Получаем актуальное состояние после upsert.
	return r.getByStudentAndType(ctx, db, sub.StudentID, sub.Type)
}

// GetActiveByStudentID возвращает все активные (не истёкшие) подписки студента.
func (r *SubscriptionRepo) GetActiveByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.Subscription, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.SubscriptionModel
	if err := db.SelectContext(ctx, &models, getActiveByStudentIDQuery, studentID); err != nil {
		return nil, errors.Wrap(err, "get active subscriptions by student id")
	}

	subs := make([]*domain.Subscription, 0, len(models))
	for i := range models {
		subs = append(subs, converter.ToDomain(&models[i]))
	}

	return subs, nil
}

// ExpireOld переводит все просроченные активные подписки в статус expired.
func (r *SubscriptionRepo) ExpireOld(ctx context.Context) error {
	db := database.GetDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, expireOldQuery); err != nil {
		return errors.Wrap(err, "expire old subscriptions")
	}

	return nil
}

// GetActiveSubscribersTelegramIDs возвращает telegram_id студентов с активной подпиской
// указанного типа. Bundle-подписка покрывает любой тип.
func (r *SubscriptionRepo) GetActiveSubscribersTelegramIDs(ctx context.Context, t domain.SubscriptionType) ([]int64, error) {
	db := database.GetDB(ctx, r.db)

	var telegramIDs []int64
	if err := db.SelectContext(ctx, &telegramIDs, getActiveSubscribersTelegramIDsQuery, string(t)); err != nil {
		return nil, errors.Wrap(err, "get active subscribers telegram ids")
	}

	return telegramIDs, nil
}

// getByStudentAndType — внутренний метод: последняя подписка студента заданного типа.
// Возвращает domain.ErrNotFound, если подписка не найдена.
func (r *SubscriptionRepo) getByStudentAndType(ctx context.Context, db database.DB, studentID uuid.UUID, subType domain.SubscriptionType) (*domain.Subscription, error) {
	var m model.SubscriptionModel
	if err := db.GetContext(ctx, &m, getByStudentAndTypeQuery, studentID, string(subType)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get subscription by student and type")
	}

	return converter.ToDomain(&m), nil
}
