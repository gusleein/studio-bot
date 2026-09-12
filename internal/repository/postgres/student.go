package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/pkg/database"
)

type studentRepo struct {
	db *sqlx.DB
}

// NewStudentRepository создаёт новый репозиторий студентов на базе PostgreSQL.
func NewStudentRepository(db *sqlx.DB) repository.StudentRepository {
	return &studentRepo{db: db}
}

func (r *studentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	var s domain.Student
	err := db.GetContext(ctx, &s,
		`SELECT id, telegram_id, username, first_name, last_name, phone, email, is_active, created_at, updated_at
		 FROM students WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("получение студента по id: %w", err)
	}
	return &s, nil
}

func (r *studentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	var s domain.Student
	err := db.GetContext(ctx, &s,
		`SELECT id, telegram_id, username, first_name, last_name, phone, email, is_active, created_at, updated_at
		 FROM students WHERE telegram_id = $1`, telegramID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("получение студента по telegram_id: %w", err)
	}
	return &s, nil
}

func (r *studentRepo) Create(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	now := time.Now()
	var result domain.Student
	err := db.GetContext(ctx, &result,
		`INSERT INTO students (id, telegram_id, username, first_name, last_name, phone, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, telegram_id, username, first_name, last_name, phone, is_active, created_at, updated_at`,
		s.ID, s.TelegramID, s.Username, s.FirstName, s.LastName, s.Phone, s.IsActive, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("создание студента: %w", err)
	}
	return &result, nil
}

func (r *studentRepo) Update(ctx context.Context, s *domain.Student) (*domain.Student, error) {
	db := database.GetDB(ctx, r.db)

	var result domain.Student
	err := db.GetContext(ctx, &result,
		`UPDATE students
		 SET username = $1, first_name = $2, last_name = $3, phone = $4, is_active = $6, updated_at = NOW()
		 WHERE id = $7
		 RETURNING id, telegram_id, username, first_name, last_name, phone, is_active, created_at, updated_at`,
		s.Username, s.FirstName, s.LastName, s.Phone, s.IsActive, s.ID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("обновление студента: %w", err)
	}
	return &result, nil
}
