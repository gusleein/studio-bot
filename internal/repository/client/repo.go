package client

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/client/converter"
	"github.com/yourstudio/studio-bot/internal/repository/client/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const clientSelect = `
		SELECT
			c.id, c.telegram_user_id, c.notes, c.is_blocked, c.last_visit,
			c.created_at, c.updated_at,
			u.id AS user_id, u.telegram_id, u.username, u.first_name, u.last_name, u.phone,
			u.created_at AS user_created_at, u.updated_at AS user_updated_at
		FROM clients c
		JOIN telegram_users u ON u.id = c.telegram_user_id`

const (
	createQuery = `
		INSERT INTO clients
			(id, telegram_user_id, notes, is_blocked, last_visit, created_at, updated_at)
		VALUES
			(:id, :telegram_user_id, :notes, :is_blocked, :last_visit, :created_at, :updated_at)`

	getByIDQuery = clientSelect + `
		WHERE c.id = $1`

	getByTelegramIDQuery = clientSelect + `
		WHERE u.telegram_id = $1`

	updateQuery = `
		UPDATE clients
		SET
			notes      = :notes,
			is_blocked = :is_blocked,
			last_visit = :last_visit,
			updated_at = :updated_at
		WHERE id = :id`
)

type Repo struct {
	db *sqlx.DB
}

var _ repository.ClientRepository = (*Repo)(nil)

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, c *domain.Client) (*domain.Client, error) {
	db := database.GetDB(ctx, r.db)

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	if _, err := db.NamedExecContext(ctx, createQuery, converter.ToModel(c)); err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	return c, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	db := database.GetDB(ctx, r.db)

	var row model.ClientWithUserRow
	if err := db.GetContext(ctx, &row, getByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get client by id")
	}
	return converter.ToDomain(&row), nil
}

func (r *Repo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Client, error) {
	db := database.GetDB(ctx, r.db)

	var row model.ClientWithUserRow
	if err := db.GetContext(ctx, &row, getByTelegramIDQuery, telegramID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get client by telegram_id")
	}
	return converter.ToDomain(&row), nil
}

func (r *Repo) Update(ctx context.Context, c *domain.Client) (*domain.Client, error) {
	db := database.GetDB(ctx, r.db)

	c.UpdatedAt = time.Now()
	res, err := db.NamedExecContext(ctx, updateQuery, converter.ToModel(c))
	if err != nil {
		return nil, errors.Wrap(err, "update client")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "update client rows affected")
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}
	return c, nil
}
