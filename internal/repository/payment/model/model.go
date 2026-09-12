package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentModel — модель платежа для хранения в БД.
type PaymentModel struct {
	ID               uuid.UUID  `db:"id"`
	StudentID        uuid.UUID  `db:"student_id"`
	SubscriptionID   *uuid.UUID `db:"subscription_id"`
	TributeUserID    int64      `db:"tribute_user_id"`
	TributeProductID int        `db:"tribute_product_id"`
	Amount           int        `db:"amount"`
	Currency         string     `db:"currency"`
	SubscriptionType string     `db:"subscription_type"`
	CreatedAt        time.Time  `db:"created_at"`
}
