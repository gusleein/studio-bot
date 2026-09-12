package payment

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/payment/converter"
	"github.com/yourstudio/studio-bot/internal/repository/payment/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const paymentColumns = `
		id, client_id, amount, currency, status,
		tribute_order_uuid, payment_url, paid_at, created_at, updated_at`

const (
	createQuery = `
		INSERT INTO payments
			(id, client_id, amount, currency, status,
			 tribute_order_uuid, payment_url, paid_at, created_at, updated_at)
		VALUES
			(:id, :client_id, :amount, :currency, :status,
			 :tribute_order_uuid, :payment_url, :paid_at, :created_at, :updated_at)`

	getByIDQuery = `
		SELECT` + paymentColumns + `
		FROM payments
		WHERE id = $1`

	getByTributeOrderUUIDQuery = `
		SELECT` + paymentColumns + `
		FROM payments
		WHERE tribute_order_uuid = $1`

	listByClientIDQuery = `
		SELECT` + paymentColumns + `
		FROM payments
		WHERE client_id = $1
		ORDER BY created_at DESC`

	updateQuery = `
		UPDATE payments
		SET
			amount             = :amount,
			currency           = :currency,
			status             = :status,
			tribute_order_uuid = :tribute_order_uuid,
			payment_url        = :payment_url,
			paid_at            = :paid_at,
			updated_at         = :updated_at
		WHERE id = :id`
)

type PaymentRepo struct {
	db *sqlx.DB
}

var _ repository.PaymentRepository = (*PaymentRepo)(nil)

func New(db *sqlx.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Currency == "" {
		p.Currency = domain.CurrencyRUB
	}
	if p.Status == "" {
		p.Status = domain.PaymentStatusPending
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	if _, err := db.NamedExecContext(ctx, createQuery, converter.ToModel(p)); err != nil {
		return nil, errors.Wrap(err, "create payment")
	}
	return p, nil
}

func (r *PaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return r.get(ctx, getByIDQuery, id, "get payment by id")
}

func (r *PaymentRepo) GetByTributeOrderUUID(ctx context.Context, orderUUID string) (*domain.Payment, error) {
	return r.get(ctx, getByTributeOrderUUIDQuery, orderUUID, "get payment by tribute order uuid")
}

func (r *PaymentRepo) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	var rows []model.PaymentModel
	if err := db.SelectContext(ctx, &rows, listByClientIDQuery, clientID); err != nil {
		return nil, errors.Wrap(err, "list payments by client_id")
	}

	result := make([]*domain.Payment, len(rows))
	for i := range rows {
		result[i] = converter.ToDomain(&rows[i])
	}
	return result, nil
}

func (r *PaymentRepo) Update(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	p.UpdatedAt = time.Now()
	res, err := db.NamedExecContext(ctx, updateQuery, converter.ToModel(p))
	if err != nil {
		return nil, errors.Wrap(err, "update payment")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "update payment rows affected")
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (r *PaymentRepo) get(ctx context.Context, query string, arg any, wrap string) (*domain.Payment, error) {
	db := database.GetDB(ctx, r.db)

	var m model.PaymentModel
	if err := db.GetContext(ctx, &m, query, arg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, wrap)
	}
	return converter.ToDomain(&m), nil
}
