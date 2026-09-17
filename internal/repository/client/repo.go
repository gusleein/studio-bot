package client

import (
	"context"
	"database/sql"
	"github.com/yourstudio/studio-bot/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/client/converter"
	"github.com/yourstudio/studio-bot/internal/repository/client/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

var _ repository.ClientRepository = (*Repo)(nil)

type Repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, c domain.Client) (result domain.Client, err error) {
	db := database.GetDB(ctx, r.db)

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err = db.NamedExecContext(ctx, `
		INSERT INTO clients
			(id, telegram_user_id, notes, is_blocked, last_visit, created_at, updated_at)
		VALUES
			(:id, :telegram_user_id, :notes, :is_blocked, :last_visit, :created_at, :updated_at)`,
		converter.ToModel(c))

	if err != nil {
		err = errors.Wrapf(err, "failed to create client with username %s", c.TgUser.Username)
	}
	return
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (result domain.Client, err error) {
	db := database.GetDB(ctx, r.db)

	var row model.ClientWithUserRow
	err = db.GetContext(ctx, &row, `
		SELECT
			c.id, c.telegram_user_id, c.notes, c.is_blocked, c.last_visit,
			c.created_at, c.updated_at,
			u.id AS user_id, u.telegram_id, u.username, u.first_name, u.last_name, u.phone,
			u.created_at AS user_created_at, u.updated_at AS user_updated_at
		FROM clients c
		JOIN telegram_users u ON u.id = c.telegram_user_id
		WHERE c.id = $1`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotFound
			return
		}
		err = errors.Wrapf(err, "get client by id %s", id)
		return
	}
	result = converter.ToDomain(row)
	return
}

func (r *Repo) GetByTelegramID(ctx context.Context, telegramID int64) (result domain.Client, err error) {
	db := database.GetDB(ctx, r.db)

	var row model.ClientWithUserRow

	err = db.GetContext(ctx, &row, `
		SELECT
			c.id, c.telegram_user_id, c.notes, c.is_blocked, c.last_visit,
			c.created_at, c.updated_at,
			u.id AS user_id, u.telegram_id, u.username, u.first_name, u.last_name, u.phone,
			u.created_at AS user_created_at, u.updated_at AS user_updated_at
		FROM clients c
		JOIN telegram_users u ON u.id = c.telegram_user_id
		WHERE u.telegram_id = $1`, telegramID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotFound
			return
		}
		err = errors.Wrapf(err, "get client by telegram id %s", telegramID)
		return
	}
	result = converter.ToDomain(row)
	return
}

func (r *Repo) Update(ctx context.Context, c domain.Client) (updated domain.Client, err error) {
	db := database.GetDB(ctx, r.db)

	c.UpdatedAt = time.Now()

	_, err = db.NamedExecContext(ctx, `
		UPDATE clients
		SET
			notes      = :notes,
			is_blocked = :is_blocked,
			last_visit = :last_visit,
			updated_at = :updated_at
		WHERE id = :id`,
		converter.ToModel(c))

	if err != nil {
		err = errors.Wrapf(err, "failed to update client with username %s", c.TgUser.Username)
		return
	}

	updated = c
	return
}
