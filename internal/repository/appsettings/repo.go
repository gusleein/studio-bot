package appsettings

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/appsettings/model"
)

var _ repository.AppSettingsRepository = (*Repo)(nil)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) GetAllAppSettings(ctx context.Context) (result []model.AppSettings, err error) {
	result = make([]model.AppSettings, 0)

	err = r.db.GetContext(ctx, &result, `SELECT * FROM app_settings`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotFound
			return
		}
		err = errors.Wrap(err, "get Settings failed")
		return
	}

	return
}

func (r *Repo) UpdateAppSettings(ctx context.Context, appSettings model.AppSettings) (err error) {
	appSettings.UpdatedAt = time.Now()

	_, err = r.db.NamedExecContext(ctx, `
		UPDATE app_settings
		SET
			value = :value,
			updated_at = :updated_at
		WHERE key = :key`, appSettings)

	return
}
