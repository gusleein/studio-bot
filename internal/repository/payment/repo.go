package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/payment/converter"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	// createPaymentQuery — вставка нового платежа.
	createPaymentQuery = `
		INSERT INTO payments
			(id, student_id, subscription_id, tribute_user_id, tribute_product_id,
			 amount, currency, subscription_type, created_at)
		VALUES
			(:id, :student_id, :subscription_id, :tribute_user_id, :tribute_product_id,
			 :amount, :currency, :subscription_type, :created_at)`

	// existsByTributeDataQuery — проверка идемпотентности: был ли уже платёж с такими данными Tribute.
	existsByTributeDataQuery = `
		SELECT COUNT(1)
		FROM payments
		WHERE tribute_user_id = $1
		  AND tribute_product_id = $2
		  AND amount = $3`
)

// PaymentRepo — репозиторий для работы с платежами.
type PaymentRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр PaymentRepo.
func New(db *sqlx.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

// Create создаёт новую запись о платеже в БД.
func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	p.ID = uuid.New()
	p.CreatedAt = time.Now()

	m := converter.ToModel(p)

	if _, err := db.NamedExecContext(ctx, createPaymentQuery, m); err != nil {
		return nil, errors.Wrap(err, "create payment")
	}

	return p, nil
}

// ExistsByTributeData проверяет, был ли уже обработан платёж с заданными данными Tribute.
// Используется для защиты от дублирующихся webhook-событий.
func (r *PaymentRepo) ExistsByTributeData(ctx context.Context, tributeUserID int64, productID int, amount int) (bool, error) {
	db := database.GetDB(ctx, r.db)

	var count int
	if err := db.GetContext(ctx, &count, existsByTributeDataQuery, tributeUserID, productID, amount); err != nil {
		return false, errors.Wrap(err, "check payment exists by tribute data")
	}

	return count > 0, nil
}
