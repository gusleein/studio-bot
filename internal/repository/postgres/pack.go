package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/database"
)

// packRow — промежуточная структура для сканирования строки из БД (поддержка pq.StringArray).
type packRow struct {
	ID             uuid.UUID       `db:"id"`
	Title          string          `db:"title"`
	Description    string          `db:"description"`
	Tags           pq.StringArray  `db:"tags"`
	FilePath       string          `db:"file_path"`
	PackType       domain.PackType `db:"pack_type"`
	TelegramFileID string          `db:"telegram_file_id"`
	Notified       bool            `db:"notified"`
	AddedAt        sql.NullTime    `db:"added_at"`
	UpdatedAt      sql.NullTime    `db:"updated_at"`
}

func (row *packRow) toDomain() *domain.Pack {
	p := &domain.Pack{
		ID:             row.ID,
		Title:          row.Title,
		Description:    row.Description,
		Tags:           []string(row.Tags),
		FilePath:       row.FilePath,
		PackType:       row.PackType,
		TelegramFileID: row.TelegramFileID,
		Notified:       row.Notified,
	}
	if row.AddedAt.Valid {
		p.AddedAt = row.AddedAt.Time
	}
	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}
	return p
}

type packRepo struct {
	db *sqlx.DB
}

// NewPackRepository создаёт новый репозиторий паков на базе PostgreSQL.
func NewPackRepository(db *sqlx.DB) repository.PackRepository {
	return &packRepo{db: db}
}

func (r *packRepo) List(ctx context.Context, packType *domain.PackType) ([]*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var rows []packRow
	var err error
	if packType != nil {
		err = db.SelectContext(ctx, &rows,
			`SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
			 FROM packs WHERE pack_type = $1 ORDER BY added_at DESC`,
			*packType,
		)
	} else {
		err = db.SelectContext(ctx, &rows,
			`SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
			 FROM packs ORDER BY added_at DESC`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("получение паков: %w", err)
	}

	result := make([]*domain.Pack, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

func (r *packRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var row packRow
	err := db.GetContext(ctx, &row,
		`SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		 FROM packs WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("получение пака по id: %w", err)
	}
	return row.toDomain(), nil
}

func (r *packRepo) Create(ctx context.Context, p *domain.Pack) (*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var row packRow
	err := db.GetContext(ctx, &row,
		`INSERT INTO packs (id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		 RETURNING id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at`,
		p.ID, p.Title, p.Description, pq.Array(p.Tags), p.FilePath, p.PackType, p.TelegramFileID, p.Notified,
	)
	if err != nil {
		return nil, fmt.Errorf("создание пака: %w", err)
	}
	return row.toDomain(), nil
}

func (r *packRepo) UpdateTelegramFileID(ctx context.Context, id uuid.UUID, fileID string) error {
	db := database.GetDB(ctx, r.db)

	_, err := db.ExecContext(ctx,
		`UPDATE packs SET telegram_file_id = $1, updated_at = NOW() WHERE id = $2`,
		fileID, id,
	)
	if err != nil {
		return fmt.Errorf("обновление telegram_file_id пака: %w", err)
	}
	return nil
}

func (r *packRepo) ListUnnotified(ctx context.Context) ([]*domain.Pack, error) {
	db := database.GetDB(ctx, r.db)

	var rows []packRow
	err := db.SelectContext(ctx, &rows,
		`SELECT id, title, description, tags, file_path, pack_type, telegram_file_id, notified, added_at, updated_at
		 FROM packs WHERE notified = FALSE ORDER BY added_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("получение неотправленных паков: %w", err)
	}

	result := make([]*domain.Pack, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

func (r *packRepo) MarkNotified(ctx context.Context, id uuid.UUID) error {
	db := database.GetDB(ctx, r.db)

	_, err := db.ExecContext(ctx,
		`UPDATE packs SET notified = TRUE, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("отметка пака как отправленного: %w", err)
	}
	return nil
}
