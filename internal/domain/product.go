package domain

import (
	"github.com/google/uuid"
	"time"
)

// Product - доп товары, типа redbull, снеки
type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64

	CreatedAt time.Time
	UpdatedAt time.Time
}
