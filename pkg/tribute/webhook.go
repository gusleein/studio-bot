package tribute

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// WebhookPayload — структура входящего вебхука от Tribute.
type WebhookPayload struct {
	Name      string         `json:"name"`
	CreatedAt string         `json:"created_at"`
	SentAt    string         `json:"sent_at"`
	Payload   ProductPayload `json:"payload"`
}

// ProductPayload — содержимое платежа внутри вебхука.
type ProductPayload struct {
	ProductID      int    `json:"product_id"`
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	UserID         int    `json:"user_id"`
	TelegramUserID int64  `json:"telegram_user_id"`
}

// VerifySignature проверяет HMAC-SHA256 подпись тела вебхука.
// signature передаётся в виде hex-строки из заголовка X-Tribute-Signature.
func VerifySignature(body []byte, signature, secretKey string) error {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("невалидная подпись вебхука")
	}

	return nil
}

// ParsePayload десериализует тело вебхука в структуру WebhookPayload.
func ParsePayload(body []byte) (*WebhookPayload, error) {
	var p WebhookPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("парсинг вебхука Tribute: %w", err)
	}

	return &p, nil
}
