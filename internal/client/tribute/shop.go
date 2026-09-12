package tribute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// API — исходящие методы Tribute Shop API, нужные боту аренды.
type API interface {
	GetShop(ctx context.Context) (*Shop, error)
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*ShopOrder, error)
	GetOrder(ctx context.Context, orderUUID string) (*ShopOrder, error)
	GetOrderStatus(ctx context.Context, orderUUID string) (*OrderStatusResponse, error)
	ListOrders(ctx context.Context, params ListOrdersParams) ([]ShopOrder, error)
	RefundTransaction(ctx context.Context, orderUUID string, txID uint64) (*RefundResponse, error)
	ResendWebhook(ctx context.Context, orderUUID, event string) (*ResendWebhookResponse, error)
}

var _ API = (*Client)(nil)

func (c *Client) GetShop(ctx context.Context) (*Shop, error) {
	var out Shop
	if err := c.do(ctx, http.MethodGet, "/shop", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*ShopOrder, error) {
	if req.ShopID == 0 && c.shopID != 0 {
		req.ShopID = c.shopID
	}
	if req.Period == "" {
		req.Period = PeriodOnetime
	}

	var out ShopOrder
	if err := c.do(ctx, http.MethodPost, "/shop/orders", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOrder(ctx context.Context, orderUUID string) (*ShopOrder, error) {
	if orderUUID == "" {
		return nil, ErrEmptyOrderUUID
	}
	var out ShopOrder
	if err := c.do(ctx, http.MethodGet, "/shop/orders/"+orderUUID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetOrderStatus(ctx context.Context, orderUUID string) (*OrderStatusResponse, error) {
	if orderUUID == "" {
		return nil, ErrEmptyOrderUUID
	}
	var out OrderStatusResponse
	if err := c.do(ctx, http.MethodGet, "/shop/orders/"+orderUUID+"/status", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListOrders(ctx context.Context, params ListOrdersParams) ([]ShopOrder, error) {
	q := url.Values{}
	if params.ShopID != 0 {
		q.Set("shopId", fmt.Sprintf("%d", params.ShopID))
	}
	if params.DateFrom != "" {
		q.Set("dateFrom", params.DateFrom)
	}
	if params.DateTo != "" {
		q.Set("dateTo", params.DateTo)
	}
	if params.PaymentLinkUUID != "" {
		q.Set("paymentLinkUuid", params.PaymentLinkUUID)
	}

	var out []ShopOrder
	if err := c.do(ctx, http.MethodGet, "/shop/orders", q, nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []ShopOrder{}
	}
	return out, nil
}

func (c *Client) RefundTransaction(ctx context.Context, orderUUID string, txID uint64) (*RefundResponse, error) {
	if orderUUID == "" {
		return nil, ErrEmptyOrderUUID
	}
	path := fmt.Sprintf("/shop/orders/%s/transactions/%d/refund", orderUUID, txID)
	var out RefundResponse
	if err := c.do(ctx, http.MethodPost, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ResendWebhook(ctx context.Context, orderUUID, event string) (*ResendWebhookResponse, error) {
	if orderUUID == "" {
		return nil, ErrEmptyOrderUUID
	}
	path := fmt.Sprintf("/shop/orders/%s/webhooks/resend", orderUUID)
	var out ResendWebhookResponse
	if err := c.do(ctx, http.MethodPost, path, nil, ResendWebhookRequest{Event: event}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
