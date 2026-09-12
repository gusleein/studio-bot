package converter

import (
	"database/sql"

	"github.com/lib/pq"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/teacher/model"
)

// ToDomain преобразует модель БД в доменный объект.
func ToDomain(m *model.TeacherModel) *domain.Teacher {
	var telegramID *int64
	if m.TelegramID.Valid {
		id := m.TelegramID.Int64
		telegramID = &id
	}

	disciplines := []string(m.Disciplines)
	if disciplines == nil {
		disciplines = []string{}
	}

	return &domain.Teacher{
		ID:          m.ID,
		Name:        m.Name,
		TelegramID:  telegramID,
		Disciplines: disciplines,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToModel преобразует доменный объект в модель БД.
func ToModel(t *domain.Teacher) *model.TeacherModel {
	var telegramID sql.NullInt64
	if t.TelegramID != nil {
		telegramID = sql.NullInt64{Int64: *t.TelegramID, Valid: true}
	}

	return &model.TeacherModel{
		ID:          t.ID,
		Name:        t.Name,
		TelegramID:  telegramID,
		Disciplines: pq.StringArray(t.Disciplines),
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
