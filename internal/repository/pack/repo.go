package pack

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/pack/converter"
	"github.com/yourstudio/studio-bot/internal/repository/pack/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	// createPackQuery — вставка нового пака.
	createPackQuery = `
		INSERT INTO packs
			(id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at)
		VALUES
			(:id, :title, :description, :tags, :file_path, :pack_type, :telegram_file_id, :notified, :added_at, :updated_at)`

	// getPackByIDQuery — пак по ID.
	getPackByIDQuery = `
		SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		FROM packs
		WHERE id = $1`

	// listAllPacksQuery — все паки, отсортированные по дате добавления.
	listAllPacksQuery = `
		SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		FROM packs
		ORDER BY added_at DESC`

	// listPacksByTypeQuery — паки заданного типа.
	listPacksByTypeQuery = `
		SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		FROM packs
		WHERE pack_type = $1
		ORDER BY added_at DESC`

	// listUnnotifiedQuery — паки, по которым ещё не было уведомлений.
	listUnnotifiedQuery = `
		SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		FROM packs
		WHERE notified = false
		ORDER BY added_at ASC`

	// markNotifiedQuery — помечает пак как уведомлённый.
	markNotifiedQuery = `
		UPDATE packs
		SET notified = true, updated_at = NOW()
		WHERE id = $1`

	// updateTelegramFileIDQuery — обновляет закешированный telegram_file_id.
	updateTelegramFileIDQuery = `
		UPDATE packs
		SET telegram_file_id = $1, updated_at = NOW()
		WHERE id = $2`
)

// PackRepo — репозиторий для работы с контент-паками.
type PackRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр PackRepo.
func New(db *sqlx.DB) *PackRepo {
	return &PackRepo{db: db}
}

// Create создаёт новый пак в БД.
func (r *PackRepo) Create(ctx context.Context, p *domain.Pack) (*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	p.ID = uuid.New()
	p.AddedAt = time.Now()
	p.UpdatedAt = time.Now()

	m := converter.ToModel(p)

	if _, err := db.NamedExecContext(ctx, createPackQuery, m); err != nil {
		return nil, errors.Wrap(err, "create pack")
	}

	return p, nil
}

// GetByID возвращает пак по UUID.
// Возвращает domain.ErrNotFound, если пак не найден.
func (r *PackRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var m model.PackModel
	if err := db.GetContext(ctx, &m, getPackByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get pack by id")
	}

	return converter.ToDomain(&m), nil
}

// List возвращает паки, отсортированные по дате добавления (новые первыми).
// Если packType == nil — возвращаются все паки; иначе — только заданного типа.
func (r *PackRepo) List(ctx context.Context, packType *domain.PackType) ([]*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.PackModel
	var err error

	if packType == nil {
		err = db.SelectContext(ctx, &models, listAllPacksQuery)
	} else {
		err = db.SelectContext(ctx, &models, listPacksByTypeQuery, string(*packType))
	}

	if err != nil {
		return nil, errors.Wrap(err, "list packs")
	}

	packs := make([]*domain.Pack, 0, len(models))
	for i := range models {
		packs = append(packs, converter.ToDomain(&models[i]))
	}

	return packs, nil
}

// ListUnnotified возвращает паки, по которым ещё не отправлялось уведомление.
func (r *PackRepo) ListUnnotified(ctx context.Context) ([]*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.PackModel
	if err := db.SelectContext(ctx, &models, listUnnotifiedQuery); err != nil {
		return nil, errors.Wrap(err, "list unnotified packs")
	}

	packs := make([]*domain.Pack, 0, len(models))
	for i := range models {
		packs = append(packs, converter.ToDomain(&models[i]))
	}

	return packs, nil
}

// MarkNotified помечает пак как уведомлённый (notified = true).
func (r *PackRepo) MarkNotified(ctx context.Context, id uuid.UUID) error {
	db := database.GetDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, markNotifiedQuery, id); err != nil {
		return errors.Wrap(err, "mark pack notified")
	}

	return nil
}

// UpdateTelegramFileID сохраняет telegram_file_id (кеш после первой отправки файла в Telegram).
func (r *PackRepo) UpdateTelegramFileID(ctx context.Context, id uuid.UUID, fileID string) error {
	db := database.GetDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, updateTelegramFileIDQuery, fileID, id); err != nil {
		return errors.Wrap(err, "update pack telegram file id")
	}

	return nil
}
