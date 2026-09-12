package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PaymentModel — строка payments.
type PaymentModel struct {
	ID               uuid.UUID    `db:"id"`
	ClientID         uuid.UUID    `db:"client_id"`
	Amount           int          `db:"amount"`
	Currency         string       `db:"currency"`
	Status           string       `db:"status"`
	TributeOrderUUID string       `db:"tribute_order_uuid"`
	PaymentURL       string       `db:"payment_url"`
	PaidAt           sql.NullTime `db:"paid_at"`
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
}
