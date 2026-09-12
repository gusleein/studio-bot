package tribute

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSignature = errors.New("невалидная подпись вебхука Tribute")
	ErrEmptyAPIKey      = errors.New("не задан API-ключ Tribute")
	ErrEmptyOrderUUID   = errors.New("не задан uuid заказа Tribute")
)

// APIError — ошибка, возвращённая Tribute Shop API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("tribute api: http %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("tribute api: http %d: %s (%s)", e.StatusCode, e.Message, e.Code)
}

type apiErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
