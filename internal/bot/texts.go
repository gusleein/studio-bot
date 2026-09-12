package bot

import (
	"fmt"
	"time"
)

// ruWeekdays — названия дней недели на русском языке.
var ruWeekdays = map[time.Weekday]string{
	time.Monday:    "Понедельник",
	time.Tuesday:   "Вторник",
	time.Wednesday: "Среда",
	time.Thursday:  "Четверг",
	time.Friday:    "Пятница",
	time.Saturday:  "Суббота",
	time.Sunday:    "Воскресенье",
}

// ruMonths — названия месяцев на русском языке (родительный падеж).
var ruMonths = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

// formatDate возвращает строку вида "Понедельник, 14 апреля".
func formatDate(t time.Time) string {
	return fmt.Sprintf("%s, %d %s", ruWeekdays[t.Weekday()], t.Day(), ruMonths[t.Month()])
}

// weekTitle возвращает HTML-заголовок недели с диапазоном дат.
// Используется в ScheduleProps.WeekTitle для шаблона schedule.botview.
func weekTitle(weekOffset int) string {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday-1)+weekOffset*7)
	sunday := monday.AddDate(0, 0, 6)

	dateRange := fmt.Sprintf("%d–%d %s", monday.Day(), sunday.Day(), ruMonths[monday.Month()])

	switch weekOffset {
	case 0:
		return fmt.Sprintf("📅 <b>Текущая неделя</b> (%s)", dateRange)
	case 1:
		return fmt.Sprintf("📅 <b>Следующая неделя</b> (%s)", dateRange)
	case -1:
		return fmt.Sprintf("📅 <b>Прошлая неделя</b> (%s)", dateRange)
	default:
		if weekOffset > 0 {
			return fmt.Sprintf("📅 <b>Через %d нед.</b> (%s)", weekOffset, dateRange)
		}
		return fmt.Sprintf("📅 <b>%d нед. назад</b> (%s)", -weekOffset, dateRange)
	}
}
