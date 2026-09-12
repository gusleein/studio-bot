package student

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/student/converter"
	"github.com/yourstudio/studio-bot/internal/repository/student/model"
	"github.com/yourstudio/studio-bot/pkg/database"
)

const (
	createStudentQuery = `
		INSERT INTO students
			(id, telegram_id, username, first_name, last_name, phone, email, is_active, created_at, updated_at)
		VALUES
			(:id, :telegram_id, :username, :first_name, :last_name, :phone, :email, :is_active, :created_at, :updated_at)`

	getStudentByTelegramIDQuery = `
		SELECT id, telegram_id, username, first_name, last_name, phone, email, is_active, created_at, updated_at
		FROM students
		WHERE telegram_id = $1`

	getStudentByIDQuery = `
		SELECT id, telegram_id, username, first_name, last_name, phone, email, is_active, created_at, updated_at
		FROM students
		WHERE id = $1`

	updateStudentQuery = `
		UPDATE students
		SET
			username   = :username,
			first_name = :first_name,
			last_name  = :last_name,
			phone      = :phone,
			email      = :email,
			is_active  = :is_active,
			updated_at = :updated_at
		WHERE id = :id`
)

// StudentRepo — репозиторий для работы с учениками.
type StudentRepo struct {
	db *sqlx.DB
}

// New создаёт новый экземпляр StudentRepo.
func New(db *sqlx.DB) *StudentRepo {
	return &StudentRepo{db: db}
}

// Create создаёт нового ученика в БД и возвращает его с заполненным ID.
func (r *StudentRepo) Create(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	m := converter.ToModel(s)

	if _, err := db.NamedExecContext(ctx, createStudentQuery, m); err != nil {
		return nil, errors.Wrap(err, "create student")
	}

	return s, nil
}

// GetByTelegramID ищет ученика по Telegram user ID.
// Возвращает domain.ErrNotFound, если ученик не найден.
func (r *StudentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	var m model.StudentModel
	if err := db.GetContext(ctx, &m, getStudentByTelegramIDQuery, telegramID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get student by telegram_id")
	}

	return converter.ToDomain(&m), nil
}

// GetByID ищет ученика по UUID.
// Возвращает domain.ErrNotFound, если ученик не найден.
func (r *StudentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	var m model.StudentModel
	if err := db.GetContext(ctx, &m, getStudentByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Wrap(err, "get student by id")
	}

	return converter.ToDomain(&m), nil
}

// Update сохраняет изменения ученика.
// Возвращает domain.ErrNotFound, если ученик не найден (RowsAffected == 0).
func (r *StudentRepo) Update(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	s.UpdatedAt = time.Now()
	m := converter.ToModel(s)

	res, err := db.NamedExecContext(ctx, updateStudentQuery, m)
	if err != nil {
		return nil, errors.Wrap(err, "update student")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "update student rows affected")
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}

	return s, nil
}
