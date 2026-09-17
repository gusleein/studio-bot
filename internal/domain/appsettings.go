package domain

type AppSettings struct {
	AppName      string        `json:"app_name"`
	OpeningHours OpeningHours  `json:"opening_hours"`
	CustomPrices []CustomPrice `json:"custom_prices"`
	BotAdmins    []BotAdmin    `json:"bot_admins"`
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

type CustomPrice struct {
	Type        string `json:"type"`
	DaysOfWeek  string `json:"days_of_week"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Price       int    `json:"price"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BotAdmin struct {
	TelegramID int      `json:"telegram_id"`
	Name       string   `json:"name"`
	Phone      string   `json:"phone"`
	Username   string   `json:"username"`
	Roles      []string `json:"roles"`
}
