package model

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionModel — модель подписки для хранения в БД.
type SubscriptionModel struct {
	ID               uuid.UUID `db:"id"`
	StudentID        uuid.UUID `db:"student_id"`
	Type             string    `db:"type"`
	Status           string    `db:"status"`
	TributeProductID int       `db:"tribute_product_id"`
	StartedAt        time.Time `db:"started_at"`
	ExpiresAt        time.Time `db:"expires_at"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
