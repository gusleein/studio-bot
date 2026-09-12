package tribute

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// VerifySignature проверяет HMAC-SHA256 тела вебхука.
// signature — значение заголовка trbt-signature (hex).
func VerifySignature(body []byte, signature, apiKey string) error {
	signature = strings.TrimSpace(signature)
	if signature == "" || apiKey == "" {
		return ErrInvalidSignature
	}

	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write(body)
	expected := mac.Sum(nil)

	got, err := hex.DecodeString(signature)
	if err != nil {
		return ErrInvalidSignature
	}
	if !hmac.Equal(expected, got) {
		return ErrInvalidSignature
	}
	return nil
}

func ParseWebhook(body []byte) (*Webhook, error) {
	var hook Webhook
	if err := json.Unmarshal(body, &hook); err != nil {
		return nil, fmt.Errorf("парсинг вебхука Tribute: %w", err)
	}
	return &hook, nil
}

func (w *Webhook) ShopOrderPayload() (*ShopOrderPayload, error) {
	return unmarshalPayload[ShopOrderPayload](w.Payload)
}

func (w *Webhook) PaymentFailedPayload() (*PaymentFailedPayload, error) {
	return unmarshalPayload[PaymentFailedPayload](w.Payload)
}

func (w *Webhook) PaymentReceivedPayload() (*PaymentReceivedPayload, error) {
	return unmarshalPayload[PaymentReceivedPayload](w.Payload)
}

func (w *Webhook) RefundedPayload() (*RefundedPayload, error) {
	return unmarshalPayload[RefundedPayload](w.Payload)
}

func unmarshalPayload[T any](raw json.RawMessage) (*T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("парсинг payload вебхука Tribute: %w", err)
	}
	return &v, nil
}
