package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/database"
)

type subscriptionRepo struct {
	db *sqlx.DB
}

// NewSubscriptionRepository создаёт новый репозиторий подписок на базе PostgreSQL.
func NewSubscriptionRepository(db *sqlx.DB) repository.SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) GetActiveByStudentID(ctx context.Context, studentID uuid.UUID) ([]*domain.Subscription, error) {
	db := database.GetDB(ctx, r.db)

	var rows []domain.Subscription
	err := db.SelectContext(ctx, &rows,
		`SELECT id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at
		 FROM subscriptions
		 WHERE student_id = $1 AND status = 'active' AND expires_at > NOW()`,
		studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("получение активных подписок студента: %w", err)
	}

	result := make([]*domain.Subscription, len(rows))
	for i := range rows {
		result[i] = &rows[i]
	}
	return result, nil
}

func (r *subscriptionRepo) GetActiveSubscribersTelegramIDs(ctx context.Context, t domain.SubscriptionType) ([]int64, error) {
	db := database.GetDB(ctx, r.db)

	var ids []int64
	err := db.SelectContext(ctx, &ids,
		`SELECT DISTINCT st.telegram_id
		 FROM subscriptions sub
		 JOIN students st ON st.id = sub.student_id
		 WHERE sub.status = 'active'
		   AND sub.expires_at > NOW()
		   AND (sub.type = $1 OR sub.type = 'bundle')`,
		t,
	)
	if err != nil {
		return nil, fmt.Errorf("получение telegram_id подписчиков: %w", err)
	}
	return ids, nil
}

func (r *subscriptionRepo) Upsert(ctx context.Context, sub *domain.Subscription) (*domain.Subscription, error) {
	db := database.GetDB(ctx, r.db)

	var result domain.Subscription
	err := db.GetContext(ctx, &result,
		`INSERT INTO subscriptions (id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, 'active', $4, $5, $6, NOW(), NOW())
		 ON CONFLICT (student_id, type) DO UPDATE SET
		     status = 'active',
		     tribute_product_id = EXCLUDED.tribute_product_id,
		     started_at = EXCLUDED.started_at,
		     expires_at = EXCLUDED.expires_at,
		     updated_at = NOW()
		 RETURNING id, student_id, type, status, tribute_product_id, started_at, expires_at, created_at, updated_at`,
		sub.ID, sub.StudentID, sub.Type, sub.TributeProductID, sub.StartedAt, sub.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert подписки: %w", err)
	}
	return &result, nil
}

func (r *subscriptionRepo) ExpireOld(ctx context.Context) error {
	db := database.GetDB(ctx, r.db)

	_, err := db.ExecContext(ctx,
		`UPDATE subscriptions
		 SET status = 'expired', updated_at = NOW()
		 WHERE status = 'active' AND expires_at < NOW()`,
	)
	if err != nil {
		return fmt.Errorf("истечение старых подписок: %w", err)
	}
	return nil
}
