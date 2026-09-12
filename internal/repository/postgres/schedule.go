package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/database"
)

type scheduleRepo struct {
	db *sqlx.DB
}

// NewScheduleRepository создаёт новый репозиторий расписания на базе PostgreSQL.
func NewScheduleRepository(db *sqlx.DB) repository.ScheduleRepository {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) List(ctx context.Context, from, to time.Time) ([]*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	var rows []domain.Schedule
	err := db.SelectContext(ctx, &rows,
		`SELECT id, course_id, teacher_id, starts_at, ends_at, room, max_students, is_cancelled, notes, sheets_row_id, created_at, updated_at
		 FROM schedule
		 WHERE starts_at >= $1 AND starts_at <= $2 AND is_cancelled = FALSE
		 ORDER BY starts_at ASC`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("получение расписания: %w", err)
	}

	result := make([]*domain.Schedule, len(rows))
	for i := range rows {
		result[i] = &rows[i]
	}
	return result, nil
}

func (r *scheduleRepo) ListByStudentID(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	var rows []domain.Schedule
	err := db.SelectContext(ctx, &rows,
		`SELECT s.id, s.course_id, s.teacher_id, s.starts_at, s.ends_at, s.room, s.max_students, s.is_cancelled, s.notes, s.sheets_row_id, s.created_at, s.updated_at
		 FROM schedule s
		 INNER JOIN schedule_students ss ON ss.schedule_id = s.id
		 WHERE ss.student_id = $1
		   AND s.starts_at >= $2 AND s.starts_at <= $3
		   AND s.is_cancelled = FALSE
		 ORDER BY s.starts_at ASC`,
		studentID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("получение расписания студента: %w", err)
	}

	result := make([]*domain.Schedule, len(rows))
	for i := range rows {
		result[i] = &rows[i]
	}
	return result, nil
}
