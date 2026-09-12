package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PackModel — модель контент-пака для хранения в БД.
type PackModel struct {
	ID             uuid.UUID      `db:"id"`
	Title          string         `db:"title"`
	Description    string         `db:"description"`
	Tags           pq.StringArray `db:"tags"`
	FilePath       string         `db:"file_path"`
	PackType       string         `db:"pack_type"`
	TelegramFileID string         `db:"telegram_file_id"`
	Notified       bool           `db:"notified"`
	AddedAt        time.Time      `db:"added_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
