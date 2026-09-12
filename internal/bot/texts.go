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

// formatClock возвращает время вида "14:00".
func formatClock(t time.Time) string {
	return t.Format("15:04")
}

// formatTimeRange возвращает интервал "14:00 – 16:00".
func formatTimeRange(start, end time.Time) string {
	return fmt.Sprintf("%s – %s", formatClock(start), formatClock(end))
}

// hoursLabel склоняет часы: 1 час, 2 часа, 5 часов.
func hoursLabel(hours int) string {
	n := hours % 100
	if n >= 11 && n <= 14 {
		return fmt.Sprintf("%d часов", hours)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%d час", hours)
	case 2, 3, 4:
		return fmt.Sprintf("%d часа", hours)
	default:
		return fmt.Sprintf("%d часов", hours)
	}
}

// formatAmount форматирует сумму в рублях.
func formatAmount(amount int) string {
	return fmt.Sprintf("%d ₽", amount)
}

// rentStatusLabel — статус брони для клиента.
func rentStatusLabel(isPaid, isCancelled bool) string {
	switch {
	case isCancelled:
		return "❌ Отменена"
	case isPaid:
		return "✅ Оплачена"
	default:
		return "⏳ Ожидает подтверждения"
	}
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
