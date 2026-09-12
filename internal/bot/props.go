package bot

// WelcomeProps — пропсы для шаблона welcome.botview.
type WelcomeProps struct {
	FirstName string
}

// RentMenuProps — главное меню аренды.
type RentMenuProps struct{}

// CallbackBtn — кнопка с текстом и callback_data.
type CallbackBtn struct {
	Callback string
	Label    string
}

// RentDatesProps — выбор даты на ближайшие 7 дней.
type RentDatesProps struct {
	Days []CallbackBtn
}

// RentHoursProps — выбор длительности.
type RentHoursProps struct {
	DateLabel string
	Hours     []CallbackBtn
}

// StartSlot — слот времени начала.
type StartSlot struct {
	Callback string
	Label    string
}

// StartSlotRow — ряд слотов в клавиатуре.
type StartSlotRow struct {
	Slots []StartSlot
}

// RentStartProps — выбор времени начала.
type RentStartProps struct {
	DateLabel  string
	HoursLabel string
	IsEmpty    bool
	Rows       []StartSlotRow
}

// RentPayProps — реквизиты для оплаты.
type RentPayProps struct {
	DateLabel  string
	TimeRange  string
	HoursLabel string
	Amount     string
	CardNumber string
}

// RentReceiptSentProps — чек отправлен, ждём админа.
type RentReceiptSentProps struct {
	DateLabel string
	TimeRange string
	Amount    string
}

// RentAdminConfirmProps — карточка для администратора.
type RentAdminConfirmProps struct {
	ClientName      string
	Username        string
	Phone           string
	DateLabel       string
	TimeRange       string
	HoursLabel      string
	Amount          string
	ConfirmCallback string
	CancelCallback  string
}

// RentConfirmedProps — подтверждение клиенту.
type RentConfirmedProps struct {
	DateLabel  string
	TimeRange  string
	HoursLabel string
}

// RentItemProps — строка аренды в списках.
type RentItemProps struct {
	DateLabel  string
	TimeRange  string
	HoursLabel string
	Amount     string
	Status     string
}

// RentUpcomingProps — ближайшая аренда.
type RentUpcomingProps struct {
	HasRent bool
	Item    RentItemProps
}

// RentHistoryProps — история аренды.
type RentHistoryProps struct {
	IsEmpty bool
	Items   []RentItemProps
}
