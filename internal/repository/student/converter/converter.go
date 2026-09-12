package converter

import (
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/student/model"
)

// ToDomain преобразует модель БД в доменный объект.
func ToDomain(m *model.StudentModel) *domain.Student {
	return &domain.Student{
		ID:         m.ID,
		TelegramID: m.TelegramID,
		Username:   m.Username,
		FirstName:  m.FirstName,
		LastName:   m.LastName,
		Phone:      m.Phone,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// ToModel преобразует доменный объект в модель БД.
func ToModel(s *domain.Student) *model.StudentModel {
	return &model.StudentModel{
		ID:         s.ID,
		TelegramID: s.TelegramID,
		Username:   s.Username,
		FirstName:  s.FirstName,
		LastName:   s.LastName,
		Phone:      s.Phone,
		IsActive:   s.IsActive,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
