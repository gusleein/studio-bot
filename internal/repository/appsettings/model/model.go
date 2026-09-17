package model

import "time"

type AppSettings struct {
	Key       string    `db:"key"`
	Value     string    `db:"value"`
	UpdatedAt time.Time `db:"updated_at"`
}

type OpeningHours struct {
	Monday    OpeningHour `json:"monday"`
	Tuesday   OpeningHour `json:"tuesday"`
	Wednesday OpeningHour `json:"wednesday"`
	Thursday  OpeningHour `json:"thursday"`
	Friday    OpeningHour `json:"friday"`
	Saturday  OpeningHour `json:"saturday"`
	Sunday    OpeningHour `json:"sunday"`
}

type OpeningHour struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type CustomPrices struct {
	Type        string `json:"type"`
	DaysOfWeek  string `json:"days_of_week"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Price       int    `json:"price"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BotAdmins struct {
	TelegramID int      `json:"telegram_id"`
	Name       string   `json:"name"`
	Phone      string   `json:"phone"`
	Username   string   `json:"username"`
	Roles      []string `json:"roles"`
}
