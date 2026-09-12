package tribute

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequiresAPIKey(t *testing.T) {
	t.Parallel()
	if _, err := New(""); err != ErrEmptyAPIKey {
		t.Fatalf("expected ErrEmptyAPIKey, got %v", err)
	}
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/shop/orders" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Api-Key"); got != "test-key" {
			t.Errorf("Api-Key = %q", got)
		}
		if r.URL.Query().Get("shopId") != "42" {
			t.Errorf("shopId query = %q", r.URL.Query().Get("shopId"))
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateOrderRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req.Amount != 10000 || req.Currency != CurrencyRUB {
			t.Errorf("request = %+v", req)
		}
		if req.Period != PeriodOnetime {
			t.Errorf("period = %s", req.Period)
		}
		if req.CustomerID != "payment-1" {
			t.Errorf("customerId = %s", req.CustomerID)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ShopOrder{
			UUID:             "550e8400-e29b-41d4-a716-446655440000",
			ShopID:           42,
			Amount:           req.Amount,
			Currency:         req.Currency,
			Title:            req.Title,
			Status:           OrderStatusPending,
			Period:           PeriodOnetime,
			WebappPaymentURL: "https://t.me/tribute/app?startapp=test",
			CustomerID:       req.CustomerID,
		})
	}))
	defer srv.Close()

	c, err := New("test-key", WithBaseURL(srv.URL), WithShopID(42))
	if err != nil {
		t.Fatal(err)
	}

	order, err := c.CreateOrder(context.Background(), CreateOrderRequest{
		Amount:      10000,
		Currency:    CurrencyRUB,
		Title:       "Аренда DJ 1ч",
		Description: "Слот 12:00",
		CustomerID:  "payment-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.UUID == "" || order.WebappPaymentURL == "" {
		t.Fatalf("incomplete order: %+v", order)
	}
	if order.Status != OrderStatusPending {
		t.Fatalf("status = %s", order.Status)
	}
}

func TestGetOrderAPIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"error_not_found","message":"order not found"}`))
	}))
	defer srv.Close()

	c, err := New("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.GetOrder(context.Background(), "missing")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "error_not_found" {
		t.Fatalf("api error = %+v", apiErr)
	}
}

func TestGetOrderEmptyUUID(t *testing.T) {
	t.Parallel()
	c, err := New("test-key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetOrder(context.Background(), ""); err != ErrEmptyOrderUUID {
		t.Fatalf("expected ErrEmptyOrderUUID, got %v", err)
	}
}

func TestVerifyAndParseWebhook(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"name":"shop_order",
		"created_at":"2025-03-20T01:15:58.33246Z",
		"sent_at":"2025-03-20T01:15:58.542279448Z",
		"payload":{"uuid":"550e8400-e29b-41d4-a716-446655440000","shopId":1,"amount":10000,"currency":"rub","fee":800,"status":"paid","customerId":"payment-1"}
	}`)

	mac := hmac.New(sha256.New, []byte("api-key"))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	if err := VerifySignature(body, sig, "api-key"); err != nil {
		t.Fatal(err)
	}
	if err := VerifySignature(body, sig, "wrong"); err != ErrInvalidSignature {
		t.Fatalf("expected invalid signature, got %v", err)
	}
	if err := VerifySignature(body, "not-hex", "api-key"); err != ErrInvalidSignature {
		t.Fatalf("expected invalid signature for bad hex, got %v", err)
	}

	hook, err := ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if hook.Name != EventShopOrder {
		t.Fatalf("name = %s", hook.Name)
	}
	payload, err := hook.ShopOrderPayload()
	if err != nil {
		t.Fatal(err)
	}
	if payload.Status != OrderStatusPaid || payload.CustomerID != "payment-1" || payload.Amount != 10000 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestCreateOrderJSONOmitsEmpty(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(CreateOrderRequest{
		Amount:      10000,
		Currency:    CurrencyRUB,
		Title:       "Rent",
		Description: "Slot",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, unexpected := range []string{"customerId", "email", "successUrl"} {
		if strings.Contains(s, unexpected) {
			t.Fatalf("unexpected field %s in %s", unexpected, s)
		}
	}
}
