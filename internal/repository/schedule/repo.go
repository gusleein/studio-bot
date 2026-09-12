package schedule

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/schedule/converter"
	"github.com/yourstudio/studio-bot/internal/repository/schedule/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	// listScheduleQuery — занятия в диапазоне дат с JOIN на таблицы teachers и courses.
	listScheduleQuery = `
		SELECT
			s.id, s.course_id, s.teacher_id, s.starts_at, s.ends_at,
			s.room, s.max_students, s.is_cancelled, s.notes, s.sheets_row_id,
			s.created_at, s.updated_at,
			COALESCE(t.name, '')  AS teacher_name,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.type, '')  AS course_type
		FROM schedule s
		LEFT JOIN teachers t ON t.id = s.teacher_id
		LEFT JOIN courses  c ON c.id = s.course_id
		WHERE s.starts_at >= $1
		  AND s.starts_at <= $2
		  AND s.is_cancelled = false
		ORDER BY s.starts_at`

	// listByStudentIDQuery — занятия студента в диапазоне дат.
	listByStudentIDQuery = `
		SELECT
			s.id, s.course_id, s.teacher_id, s.starts_at, s.ends_at,
			s.room, s.max_students, s.is_cancelled, s.notes, s.sheets_row_id,
			s.created_at, s.updated_at,
			COALESCE(t.name, '')  AS teacher_name,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.type, '')  AS course_type
		FROM schedule s
		LEFT JOIN teachers        t  ON t.id  = s.teacher_id
		LEFT JOIN courses         c  ON c.id  = s.course_id
		JOIN      schedule_students ss ON ss.schedule_id = s.id
		WHERE ss.student_id = $1
		  AND s.starts_at >= $2
		  AND s.starts_at <= $3
		  AND s.is_cancelled = false
		ORDER BY s.starts_at`

	// createScheduleQuery — вставка нового занятия.
	createScheduleQuery = `
		INSERT INTO schedule
			(id, course_id, teacher_id, starts_at, ends_at, room, max_students,
			 is_cancelled, notes, sheets_row_id, created_at, updated_at)
		VALUES
			(:id, :course_id, :teacher_id, :starts_at, :ends_at, :room, :max_students,
			 :is_cancelled, :notes, :sheets_row_id, :created_at, :updated_at)`

	// getScheduleByIDQuery — занятие по ID с JOIN.
	getScheduleByIDQuery = `
		SELECT
			s.id, s.course_id, s.teacher_id, s.starts_at, s.ends_at,
			s.room, s.max_students, s.is_cancelled, s.notes, s.sheets_row_id,
			s.created_at, s.updated_at,
			COALESCE(t.name, '')  AS teacher_name,
			COALESCE(c.title, '') AS course_title,
			COALESCE(c.type, '')  AS course_type
		FROM schedule s
		LEFT JOIN teachers t ON t.id = s.teacher_id
		LEFT JOIN courses  c ON c.id = s.course_id
		WHERE s.id = $1`

	// addStudentQuery — запись студента на занятие; дубли игнорируются.
	addStudentQuery = `
		INSERT INTO schedule_students (schedule_id, student_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`
)

// ScheduleRepo — репозиторий для работы с расписанием.
type ScheduleRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр ScheduleRepo.
func New(db *sqlx.DB) *ScheduleRepo {
	return &ScheduleRepo{db: db}
}

// List возвращает не отменённые занятия в диапазоне дат [from, to].
func (r *ScheduleRepo) List(ctx context.Context, from, to time.Time) ([]*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.ScheduleModel
	if err := db.SelectContext(ctx, &models, listScheduleQuery, from, to); err != nil {
		return nil, errors.Wrap(err, "list schedule")
	}

	schedules := make([]*domain.Schedule, 0, len(models))
	for i := range models {
		schedules = append(schedules, converter.ToDomain(&models[i]))
	}

	return schedules, nil
}

// ListByStudentID возвращает занятия, на которые записан студент, в заданном диапазоне дат.
func (r *ScheduleRepo) ListByStudentID(ctx context.Context, studentID uuid.UUID, from, to time.Time) ([]*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.ScheduleModel
	if err := db.SelectContext(ctx, &models, listByStudentIDQuery, studentID, from, to); err != nil {
		return nil, errors.Wrap(err, "list schedule by student id")
	}

	schedules := make([]*domain.Schedule, 0, len(models))
	for i := range models {
		schedules = append(schedules, converter.ToDomain(&models[i]))
	}

	return schedules, nil
}

// Create создаёт новое занятие в расписании.
func (r *ScheduleRepo) Create(ctx context.Context, s *domain.Schedule) (*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	m := converter.ToModel(s)

	if _, err := db.NamedExecContext(ctx, createScheduleQuery, m); err != nil {
		return nil, errors.Wrap(err, "create schedule")
	}

	return s, nil
}

// GetByID возвращает занятие по UUID (с JOIN на teachers и courses).
// Возвращает domain.ErrNotFound, если занятие не найдено.
func (r *ScheduleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	db := database.GetDB(ctx, r.db)

	var m model.ScheduleModel
	if err := db.GetContext(ctx, &m, getScheduleByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get schedule by id")
	}

	return converter.ToDomain(&m), nil
}

// AddStudent записывает студента на занятие.
// Если студент уже записан (ON CONFLICT DO NOTHING), ошибки не возникает.
func (r *ScheduleRepo) AddStudent(ctx context.Context, ss *domain.ScheduleStudent) (*domain.ScheduleStudent, error) {
	db := database.GetDB(ctx, r.db)

	ss.CreatedAt = time.Now()

	if _, err := db.ExecContext(ctx, addStudentQuery, ss.ScheduleID, ss.StudentID, ss.CreatedAt); err != nil {
		return nil, errors.Wrap(err, "add student to schedule")
	}

	return ss, nil
}
