package converter

import (
	"database/sql"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/payment/model"
)

func ToDomain(m *model.PaymentModel) *domain.Payment {
	p := &domain.Payment{
		ID:               m.ID,
		ClientID:         m.ClientID,
		Amount:           m.Amount,
		Currency:         domain.Currency(m.Currency),
		Status:           domain.PaymentStatus(m.Status),
		TributeOrderUUID: m.TributeOrderUUID,
		PaymentURL:       m.PaymentURL,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.PaidAt.Valid {
		t := m.PaidAt.Time
		p.PaidAt = &t
	}
	return p
}

func ToModel(p *domain.Payment) *model.PaymentModel {
	m := &model.PaymentModel{
		ID:               p.ID,
		ClientID:         p.ClientID,
		Amount:           p.Amount,
		Currency:         string(p.Currency),
		Status:           string(p.Status),
		TributeOrderUUID: p.TributeOrderUUID,
		PaymentURL:       p.PaymentURL,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
	if p.PaidAt != nil {
		m.PaidAt = sql.NullTime{Time: *p.PaidAt, Valid: true}
	}
	return m
}
