package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/database"
)

type paymentRepo struct {
	db *sqlx.DB
}

// NewPaymentRepository создаёт новый репозиторий платежей на базе PostgreSQL.
func NewPaymentRepository(db *sqlx.DB) repository.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	var result domain.Payment
	err := db.GetContext(ctx, &result,
		`INSERT INTO payments (id, student_id, subscription_id, tribute_user_id, tribute_product_id, amount, currency, subscription_type, created_at)
		 VALUES ($1, $2, $3, $4, $5,  NOW())
		 RETURNING id, student_id, subscription_id, tribute_user_id, tribute_product_id, amount, currency, subscription_type, created_at`,
		p.ID, p.TributeUserID, p.TributeProductID, p.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("создание платежа: %w", err)
	}
	return &result, nil
}

func (r *paymentRepo) ExistsByTributeData(ctx context.Context, tributeUserID int64, tributeProductID int, amount int) (bool, error) {
	db := database.GetDB(ctx, r.db)

	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(1) FROM payments WHERE tribute_user_id = $1 AND tribute_product_id = $2 AND amount = $3`,
		tributeUserID, tributeProductID, amount,
	)
	if err != nil {
		return false, fmt.Errorf("проверка существования платежа: %w", err)
	}
	return count > 0, nil
}
