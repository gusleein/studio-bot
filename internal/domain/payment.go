package domain

import (
	"github.com/google/uuid"
	"time"
)

// Payment — запись об успешном платеже через Tribute.
// Для каждого вида товаров будет
type Payment struct {
	ID uuid.UUID

	TributeUserID    int64
	TributeProductID int

	Amount int

	CreatedAt time.Time
}
