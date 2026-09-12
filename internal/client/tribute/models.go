package tribute

import (
	"encoding/json"
	"time"
)

type Currency string

const (
	CurrencyRUB Currency = "rub"
	CurrencyUSD Currency = "usd"
	CurrencyEUR Currency = "eur"
)

type Period string

const (
	PeriodOnetime    Period = "onetime"
	PeriodWeekly     Period = "weekly"
	PeriodMonthly    Period = "monthly"
	PeriodQuarterly  Period = "quarterly"
	PeriodHalfYearly Period = "halfyearly"
	PeriodYearly     Period = "yearly"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusPrepaid OrderStatus = "prepaid"
	OrderStatusPaid    OrderStatus = "paid"
	OrderStatusFailed  OrderStatus = "failed"
)

const (
	EventShopOrder       = "shop_order"
	EventPaymentReceived = "shop_order_payment_received"
	EventPaymentFailed   = "shop_order_payment_failed"
	EventOrderRefunded   = "shop_order_refunded"
	EventChargeSuccess   = "shop_order_charge_success"
	EventChargeFailed    = "shop_order_charge_failed"
	EventOrderCancelled  = "shop_order_cancelled"
	SignatureHeader      = "trbt-signature"
)

// CreateOrderRequest — тело POST /shop/orders.
type CreateOrderRequest struct {
	ShopID      uint64   `json:"shopId,omitempty"`
	Amount      int64    `json:"amount"`
	Currency    Currency `json:"currency"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CustomerID  string   `json:"customerId,omitempty"`
	Period      Period   `json:"period,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Email       string   `json:"email,omitempty"`
	SuccessURL  string   `json:"successUrl,omitempty"`
	FailURL     string   `json:"failUrl,omitempty"`
}

// ShopOrder — заказ магазина Tribute.
type ShopOrder struct {
	UUID             string      `json:"uuid"`
	ShopID           uint64      `json:"shopId"`
	Amount           int64       `json:"amount"`
	Currency         Currency    `json:"currency"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Status           OrderStatus `json:"status"`
	Email            string      `json:"email,omitempty"`
	SuccessURL       string      `json:"successUrl"`
	FailURL          string      `json:"failUrl"`
	PaymentURL       *string     `json:"paymentUrl"`
	WebappPaymentURL string      `json:"webappPaymentUrl"`
	CreatedAt        time.Time   `json:"createdAt"`
	Comment          string      `json:"comment,omitempty"`
	Period           Period      `json:"period"`
	CustomerID       string      `json:"customerId,omitempty"`
	PaymentLinkUUID  string      `json:"paymentLinkUuid,omitempty"`
}

// OrderStatusResponse — тело GET /shop/orders/{uuid}/status.
type OrderStatusResponse struct {
	Status OrderStatus `json:"status"`
}

// Shop — магазин Tribute.
type Shop struct {
	ID            uint64 `json:"id"`
	UserID        int64  `json:"userId"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	CallbackURL   string `json:"callbackUrl"`
	Recurrent     bool   `json:"recurrent"`
	OnlyStars     bool   `json:"onlyStars"`
	TokenCharging bool   `json:"tokenCharging"`
	Status        int    `json:"status"`
}

// ListOrdersParams — query для GET /shop/orders.
type ListOrdersParams struct {
	ShopID          uint64
	DateFrom        string // yyyy-mm-dd
	DateTo          string
	PaymentLinkUUID string
}

// RefundResponse — тело POST .../transactions/{txId}/refund.
type RefundResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ResendWebhookRequest — тело POST .../webhooks/resend.
type ResendWebhookRequest struct {
	Event string `json:"event"`
}

type ResendWebhookResponse struct {
	Success bool `json:"success"`
}

// Webhook — конверт входящего вебхука магазина.
type Webhook struct {
	Name      string          `json:"name"`
	CreatedAt time.Time       `json:"created_at"`
	SentAt    time.Time       `json:"sent_at"`
	Payload   json.RawMessage `json:"payload"`
}

// ShopOrderPayload — payload события shop_order.
type ShopOrderPayload struct {
	UUID       string      `json:"uuid"`
	ShopID     uint64      `json:"shopId"`
	Amount     int64       `json:"amount"`
	Currency   Currency    `json:"currency"`
	Fee        int64       `json:"fee"`
	Status     OrderStatus `json:"status"`
	CustomerID string      `json:"customerId,omitempty"`
	IsTrial    bool        `json:"isTrial,omitempty"`
}

// PaymentFailedPayload — payload shop_order_payment_failed.
type PaymentFailedPayload struct {
	UUID         string   `json:"uuid"`
	ShopID       uint64   `json:"shopId"`
	Amount       int64    `json:"amount"`
	Currency     Currency `json:"currency"`
	ErrorCode    string   `json:"errorCode"`
	ErrorMessage string   `json:"errorMessage"`
	CustomerID   string   `json:"customerId,omitempty"`
}

// PaymentReceivedPayload — payload shop_order_payment_received (ещё не финальная оплата).
type PaymentReceivedPayload struct {
	UUID       string   `json:"uuid"`
	ShopID     uint64   `json:"shopId"`
	Amount     int64    `json:"amount"`
	Currency   Currency `json:"currency"`
	CustomerID string   `json:"customerId,omitempty"`
}

// RefundedPayload — payload shop_order_refunded.
type RefundedPayload struct {
	UUID          string    `json:"uuid"`
	ShopID        uint64    `json:"shopId"`
	TransactionID uint64    `json:"transactionId"`
	Amount        int64     `json:"amount"`
	Currency      Currency  `json:"currency"`
	Status        string    `json:"status"`
	RefundedAt    time.Time `json:"refundedAt,omitempty"`
	CustomerID    string    `json:"customerId,omitempty"`
}
