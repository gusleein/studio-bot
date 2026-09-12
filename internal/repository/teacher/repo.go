package teacher

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/teacher/converter"
	"github.com/yourstudio/studio-bot/internal/repository/teacher/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	createTeacherQuery = `
		INSERT INTO teachers
			(id, name, telegram_id, disciplines, is_active, created_at, updated_at)
		VALUES
			(:id, :name, :telegram_id, :disciplines, :is_active, :created_at, :updated_at)`

	getAllTeachersQuery = `
		SELECT id, name, telegram_id, disciplines, is_active, created_at, updated_at
		FROM teachers
		WHERE is_active = true
		ORDER BY name`

	getTeacherByIDQuery = `
		SELECT id, name, telegram_id, disciplines, is_active, created_at, updated_at
		FROM teachers
		WHERE id = $1`
)

// TeacherRepo — репозиторий для работы с преподавателями.
type TeacherRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр TeacherRepo.
func New(db *sqlx.DB) *TeacherRepo {
	return &TeacherRepo{db: db}
}

// Create создаёт нового преподавателя в БД.
func (r *TeacherRepo) Create(ctx context.Context, t *domain.Teacher) (*domain.Teacher, error) {
	db := database.GetDB(ctx, r.db)

	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	m := converter.ToModel(t)

	if _, err := db.NamedExecContext(ctx, createTeacherQuery, m); err != nil {
		return nil, errors.Wrap(err, "create teacher")
	}

	return t, nil
}

// GetAll возвращает всех активных преподавателей, отсортированных по имени.
func (r *TeacherRepo) GetAll(ctx context.Context) ([]*domain.Teacher, error) {
	db := database.GetDB(ctx, r.db)

	var models []model.TeacherModel
	if err := db.SelectContext(ctx, &models, getAllTeachersQuery); err != nil {
		return nil, errors.Wrap(err, "get all teachers")
	}

	teachers := make([]*domain.Teacher, 0, len(models))
	for i := range models {
		teachers = append(teachers, converter.ToDomain(&models[i]))
	}

	return teachers, nil
}

// GetByID ищет преподавателя по UUID.
// Возвращает domain.ErrNotFound, если преподаватель не найден.
func (r *TeacherRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Teacher, error) {
	db := database.GetDB(ctx, r.db)

	var m model.TeacherModel
	if err := db.GetContext(ctx, &m, getTeacherByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get teacher by id")
	}

	return converter.ToDomain(&m), nil
}
