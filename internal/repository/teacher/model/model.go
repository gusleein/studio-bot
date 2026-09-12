package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TeacherModel — модель преподавателя для хранения в БД.
type TeacherModel struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	TelegramID  sql.NullInt64  `db:"telegram_id"`
	Disciplines pq.StringArray `db:"disciplines"`
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
