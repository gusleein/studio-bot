package converter

import (
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/telegramuser/model"
)

func ToDomain(m model.TelegramUserModel) domain.TelegramUser {
	return domain.TelegramUser{
		Id:         m.ID,
		TelegramId: m.TelegramID,
		Username:   m.Username,
		FirstName:  m.FirstName,
		LastName:   m.LastName,
		Phone:      m.Phone,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func ToModel(u domain.TelegramUser) model.TelegramUserModel {
	return model.TelegramUserModel{
		ID:         u.Id,
		TelegramID: u.TelegramId,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Phone:      u.Phone,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}
