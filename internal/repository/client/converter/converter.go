package converter

import (
	"database/sql"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/client/model"
)

func ToDomain(row *model.ClientWithUserRow) *domain.Client {
	c := &domain.Client{
		ID: row.ID,
		TgUser: domain.TelegramUser{
			Id:         row.UserID,
			TelegramId: row.TelegramID,
			Username:   row.Username,
			FirstName:  row.FirstName,
			LastName:   row.LastName,
			Phone:      row.Phone,
			CreatedAt:  row.UserCreatedAt,
			UpdatedAt:  row.UserUpdatedAt,
		},
		Notes:     row.Notes,
		IsBlocked: row.IsBlocked,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.LastVisit.Valid {
		c.LastVisit = row.LastVisit.Time
	}
	return c
}

func ToModel(c *domain.Client) *model.ClientModel {
	m := &model.ClientModel{
		ID:             c.ID,
		TelegramUserID: c.TgUser.Id,
		Notes:          c.Notes,
		IsBlocked:      c.IsBlocked,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
	if !c.LastVisit.IsZero() {
		m.LastVisit = sql.NullTime{Time: c.LastVisit, Valid: true}
	}
	return m
}
