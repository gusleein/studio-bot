package domain

import (
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	CurrencyRUB Currency = "rub"
	CurrencyUSD Currency = "usd"
	CurrencyEUR Currency = "eur"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// Payment — попытка оплаты у нас. Сумма всегда в копейках/центах.
type Payment struct {
	ID       uuid.UUID
	ClientID uuid.UUID
	Amount   int // 100 ₽ = 10000
	Currency Currency
	Status   PaymentStatus

	TributeOrderUUID string // uuid заказа Shop API
	PaymentURL       string // webappPaymentUrl, отдали клиенту в боте

	PaidAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
