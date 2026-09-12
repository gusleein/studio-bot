package converter

import (
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/subscription/model"
)

// ToDomain преобразует модель БД в доменный объект Subscription.
func ToDomain(m *model.SubscriptionModel) *domain.Subscription {
	return &domain.Subscription{
		ID:               m.ID,
		StudentID:        m.StudentID,
		Type:             domain.SubscriptionType(m.Type),
		Status:           domain.SubscriptionStatus(m.Status),
		TributeProductID: m.TributeProductID,
		StartedAt:        m.StartedAt,
		ExpiresAt:        m.ExpiresAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

// ToModel преобразует доменный объект Subscription в модель БД.
func ToModel(s *domain.Subscription) *model.SubscriptionModel {
	return &model.SubscriptionModel{
		ID:               s.ID,
		StudentID:        s.StudentID,
		Type:             string(s.Type),
		Status:           string(s.Status),
		TributeProductID: s.TributeProductID,
		StartedAt:        s.StartedAt,
		ExpiresAt:        s.ExpiresAt,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
