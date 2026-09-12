package telegramuser

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/telegramuser/converter"
	"github.com/yourstudio/studio-bot/internal/repository/telegramuser/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	createQuery = `
		INSERT INTO telegram_users
			(id, telegram_id, username, first_name, last_name, phone, created_at, updated_at)
		VALUES
			(:id, :telegram_id, :username, :first_name, :last_name, :phone, :created_at, :updated_at)`

	getByIDQuery = `
		SELECT id, telegram_id, username, first_name, last_name, phone, created_at, updated_at
		FROM telegram_users
		WHERE id = $1`

	getByTelegramIDQuery = `
		SELECT id, telegram_id, username, first_name, last_name, phone, created_at, updated_at
		FROM telegram_users
		WHERE telegram_id = $1`

	updateQuery = `
		UPDATE telegram_users
		SET
			username   = :username,
			first_name = :first_name,
			last_name  = :last_name,
			phone      = :phone,
			updated_at = :updated_at
		WHERE id = :id`
)

type Repo struct {
	db *sqlx.DB
}

var _ repository.TelegramUserRepository = (*Repo)(nil)

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, u *domain.TelegramUser) (*domain.TelegramUser, error) {
	db := database.GetDB(ctx, r.db)

	if u.Id == uuid.Nil {
		u.Id = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	if _, err := db.NamedExecContext(ctx, createQuery, converter.ToModel(u)); err != nil {
		return nil, errors.Wrap(err, "create telegram user")
	}
	return u, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.TelegramUser, error) {
	db := database.GetDB(ctx, r.db)

	var m model.TelegramUserModel
	if err := db.GetContext(ctx, &m, getByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get telegram user by id")
	}
	return converter.ToDomain(&m), nil
}

func (r *Repo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramUser, error) {
	db := database.GetDB(ctx, r.db)

	var m model.TelegramUserModel
	if err := db.GetContext(ctx, &m, getByTelegramIDQuery, telegramID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get telegram user by telegram_id")
	}
	return converter.ToDomain(&m), nil
}

func (r *Repo) Update(ctx context.Context, u *domain.TelegramUser) (*domain.TelegramUser, error) {
	db := database.GetDB(ctx, r.db)

	u.UpdatedAt = time.Now()
	res, err := db.NamedExecContext(ctx, updateQuery, converter.ToModel(u))
	if err != nil {
		return nil, errors.Wrap(err, "update telegram user")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "update telegram user rows affected")
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}
	return u, nil
}
