package rent

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/rent/converter"
	"github.com/yourstudio/studio-bot/internal/repository/rent/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const rentColumns = `
		id, client_id, starts_at, ends_at, is_cancelled, is_paid,
		price_per_hour, paid_duration, bonus_duration, door_code, type,
		paid_at, notes, created_at, updated_at`

const (
	createQuery = `
		INSERT INTO rents
			(id, client_id, starts_at, ends_at, is_cancelled, is_paid,
			 price_per_hour, paid_duration, bonus_duration, door_code, type,
			 paid_at, notes, created_at, updated_at)
		VALUES
			(:id, :client_id, :starts_at, :ends_at, :is_cancelled, :is_paid,
			 :price_per_hour, :paid_duration, :bonus_duration, :door_code, :type,
			 :paid_at, :notes, :created_at, :updated_at)`

	getByIDQuery = `
		SELECT` + rentColumns + `
		FROM rents
		WHERE id = $1`

	listByClientIDQuery = `
		SELECT` + rentColumns + `
		FROM rents
		WHERE client_id = $1
		ORDER BY starts_at DESC`

	updateQuery = `
		UPDATE rents
		SET
			starts_at      = :starts_at,
			ends_at        = :ends_at,
			is_cancelled   = :is_cancelled,
			is_paid        = :is_paid,
			price_per_hour = :price_per_hour,
			paid_duration  = :paid_duration,
			bonus_duration = :bonus_duration,
			door_code      = :door_code,
			type           = :type,
			paid_at        = :paid_at,
			notes          = :notes,
			updated_at     = :updated_at
		WHERE id = :id`
)

type Repo struct {
	db *sqlx.DB
}

var _ repository.RentRepository = (*Repo)(nil)

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, rent *domain.Rent) (*domain.Rent, error) {
	db := database.GetDB(ctx, r.db)

	if rent.ID == uuid.Nil {
		rent.ID = uuid.New()
	}
	now := time.Now()
	rent.CreatedAt = now
	rent.UpdatedAt = now

	if _, err := db.NamedExecContext(ctx, createQuery, converter.ToModel(rent)); err != nil {
		return nil, errors.Wrap(err, "create rent")
	}
	return rent, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rent, error) {
	db := database.GetDB(ctx, r.db)

	var m model.RentModel
	if err := db.GetContext(ctx, &m, getByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get rent by id")
	}
	return converter.ToDomain(&m), nil
}

func (r *Repo) ListByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Rent, error) {
	db := database.GetDB(ctx, r.db)

	var rows []model.RentModel
	if err := db.SelectContext(ctx, &rows, listByClientIDQuery, clientID); err != nil {
		return nil, errors.Wrap(err, "list rents by client_id")
	}

	result := make([]*domain.Rent, len(rows))
	for i := range rows {
		result[i] = converter.ToDomain(&rows[i])
	}
	return result, nil
}

func (r *Repo) Update(ctx context.Context, rent *domain.Rent) (*domain.Rent, error) {
	db := database.GetDB(ctx, r.db)

	rent.UpdatedAt = time.Now()
	res, err := db.NamedExecContext(ctx, updateQuery, converter.ToModel(rent))
	if err != nil {
		return nil, errors.Wrap(err, "update rent")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "update rent rows affected")
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}
	return rent, nil
}
