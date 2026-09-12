package converter

import (
	"database/sql"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/rent/model"
)

func ToDomain(m *model.RentModel) *domain.Rent {
	r := &domain.Rent{
		ID:            m.ID,
		ClientID:      m.ClientID,
		StartsAt:      m.StartsAt,
		IsCancelled:   m.IsCancelled,
		IsPaid:        m.IsPaid,
		PricePerHour:  m.PricePerHour,
		PaidDuration:  m.PaidDuration,
		BonusDuration: m.BonusDuration,
		DoorCode:      m.DoorCode,
		Type:          m.Type,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.EndsAt.Valid {
		r.EndsAt = m.EndsAt.Time
	}
	if m.PaidAt.Valid {
		r.PaidAt = m.PaidAt.Time
	}
	return r
}

func ToModel(r *domain.Rent) *model.RentModel {
	m := &model.RentModel{
		ID:            r.ID,
		ClientID:      r.ClientID,
		StartsAt:      r.StartsAt,
		IsCancelled:   r.IsCancelled,
		IsPaid:        r.IsPaid,
		PricePerHour:  r.PricePerHour,
		PaidDuration:  r.PaidDuration,
		BonusDuration: r.BonusDuration,
		DoorCode:      r.DoorCode,
		Type:          r.Type,
		Notes:         r.Notes,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if !r.EndsAt.IsZero() {
		m.EndsAt = sql.NullTime{Time: r.EndsAt, Valid: true}
	}
	if !r.PaidAt.IsZero() {
		m.PaidAt = sql.NullTime{Time: r.PaidAt, Valid: true}
	}
	return m
}
