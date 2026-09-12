package converter

import (
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/payment/model"
)

// ToDomain преобразует модель БД в доменный объект Payment.
func ToDomain(m *model.PaymentModel) *domain.Payment {
	return &domain.Payment{
		ID:               m.ID,
		TributeUserID:    m.TributeUserID,
		TributeProductID: m.TributeProductID,
		Amount:           m.Amount,
		CreatedAt:        m.CreatedAt,
	}
}

// ToModel преобразует доменный объект Payment в модель БД.
func ToModel(p *domain.Payment) *model.PaymentModel {
	return &model.PaymentModel{
		ID:               p.ID,
		TributeUserID:    p.TributeUserID,
		TributeProductID: p.TributeProductID,
		Amount:           p.Amount,
		CreatedAt:        p.CreatedAt,
	}
}
