package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
)

const (
	adminRentType  = "admin"
	adminDaysAhead = 14 // на сколько дней вперёд админ видит и может бронировать
)

// handleAdmin — точка входа команды /admin.
// Сразу показывает предстоящие брони текущего администратора.
func (h *Handler) handleAdmin(ctx context.Context, msg *tgbotapi.Message) {
	if _, ok := h.requireAdmin(ctx, msg.From, msg.Chat.ID); !ok {
		return
	}

	client, ok := h.requireClient(ctx, msg.From, msg.Chat.ID)
	if !ok {
		return
	}

	h.showAdminSchedule(ctx, client, msg.Chat.ID, 0, false)
}

// handleAdminRentCallback обрабатывает callback'и админского раздела бронирования.
// payload — часть после префикса "adminrent:".
func (h *Handler) handleAdminRentCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, payload string) {
	if _, ok := h.requireAdmin(ctx, cq.From, cq.Message.Chat.ID); !ok {
		return
	}
	client, ok := h.requireClient(ctx, cq.From, cq.Message.Chat.ID)
	if !ok {
		return
	}

	chatID := cq.Message.Chat.ID
	msgID := cq.Message.MessageID

	switch {
	case payload == "menu":
		h.showAdminSchedule(ctx, client, chatID, msgID, false)
	case payload == "all":
		h.showAdminSchedule(ctx, client, chatID, msgID, true)
	case payload == "book":
		h.startAdminBooking(ctx, client, chatID, msgID)
	case payload == "hours_back":
		h.showAdminHours(client.TgUser.TelegramId, chatID, msgID)
	case strings.HasPrefix(payload, "date:"):
		h.pickAdminDate(client.TgUser.TelegramId, chatID, msgID, strings.TrimPrefix(payload, "date:"))
	case strings.HasPrefix(payload, "hours:"):
		h.pickAdminHours(ctx, client.TgUser.TelegramId, chatID, msgID, strings.TrimPrefix(payload, "hours:"))
	case strings.HasPrefix(payload, "start:"):
		h.pickAdminStart(ctx, client, chatID, msgID, strings.TrimPrefix(payload, "start:"))
	default:
		h.log.Warn("неизвестный adminrent callback", zap.String("payload", payload))
	}
}

// requireAdmin проверяет, что пользователь есть в appSettings.BotAdmins.
func (h *Handler) requireAdmin(ctx context.Context, from *tgbotapi.User, chatID int64) (domain.BotAdmin, bool) {
	if from == nil {
		return domain.BotAdmin{}, false
	}

	settings, err := h.appSettings.Get(ctx)
	if err != nil {
		h.log.Error("ошибка загрузки настроек приложения", zap.Error(err))
		h.send(chatID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return domain.BotAdmin{}, false
	}

	for _, admin := range settings.BotAdmins {
		if int64(admin.TelegramID) == from.ID {
			return admin, true
		}
	}

	h.send(chatID, "⛔ У вас нет доступа к этому разделу.", nil)
	return domain.BotAdmin{}, false
}

// showAdminSchedule показывает предстоящие брони:
// свои (isTeamView == false) или брони всей команды (isTeamView == true) на adminDaysAhead дней вперёд.
func (h *Handler) showAdminSchedule(ctx context.Context, client *domain.Client, chatID int64, messageID int, isTeamView bool) {
	from := time.Now().In(h.loc)
	to := from.AddDate(0, 0, adminDaysAhead)

	var rents []*domain.Rent
	var err error
	if isTeamView {
		rents, err = h.rents.UpcomingAll(ctx, from, to)
	} else {
		rents, err = h.rents.UpcomingByClient(ctx, client.ID, from, to)
	}
	if err != nil {
		h.log.Error("ошибка загрузки расписания администратора", zap.Bool("team_view", isTeamView), zap.Error(err))
		h.send(chatID, "⚠️ Не удалось загрузить расписание.", nil)
		return
	}

	props := AdminScheduleProps{
		IsTeamView: isTeamView,
		IsEmpty:    len(rents) == 0,
	}
	for _, r := range rents {
		item := AdminRentItemProps{
			DateLabel:  formatDate(r.StartsAt.In(h.loc)),
			TimeRange:  formatTimeRange(r.StartsAt.In(h.loc), r.EndsAt.In(h.loc)),
			HoursLabel: hoursLabel(r.PaidDuration),
		}
		if isTeamView {
			item.BookedBy = h.rentOwnerLabel(ctx, r.ClientID, client.ID)
		}
		props.Items = append(props.Items, item)
	}

	if messageID > 0 {
		h.renderAndEdit(chatID, messageID, "admin_schedule", props)
		return
	}
	h.renderAndSend(chatID, "admin_schedule", props)
}

// rentOwnerLabel определяет подпись автора брони для общего расписания команды.
func (h *Handler) rentOwnerLabel(ctx context.Context, rentClientID, currentClientID uuid.UUID) string {
	if rentClientID == currentClientID {
		return "Вы"
	}
	owner, err := h.clients.GetByID(ctx, rentClientID)
	if err != nil {
		h.log.Warn("не удалось определить владельца брони", zap.Error(err))
		return "—"
	}
	name := strings.TrimSpace(owner.TgUser.FirstName + " " + owner.TgUser.LastName)
	if name == "" {
		name = owner.TgUser.Username
	}
	if name == "" {
		name = "—"
	}
	return name
}

func (h *Handler) startAdminBooking(ctx context.Context, client *domain.Client, chatID int64, messageID int) {
	if sess := h.sessions.get(client.TgUser.TelegramId); sess != nil && sess.RentID != uuid.Nil {
		_ = h.rents.CancelUnpaid(ctx, sess.RentID)
	}

	h.sessions.put(client.TgUser.TelegramId, &bookingSession{
		ClientID: client.ID,
		ChatID:   chatID,
	})
	h.showAdminDates(chatID, messageID)
}

func (h *Handler) showAdminDates(chatID int64, messageID int) {
	now := time.Now().In(h.loc)
	days := make([]CallbackBtn, 0, adminDaysAhead)
	for i := 0; i < adminDaysAhead; i++ {
		d := now.AddDate(0, 0, i)
		label := formatDate(d)
		if i == 0 {
			label = "Сегодня, " + strconv.Itoa(d.Day()) + " " + ruMonths[d.Month()]
		}
		days = append(days, CallbackBtn{
			Callback: "adminrent:date:" + d.Format("2006-01-02"),
			Label:    label,
		})
	}
	h.renderAndEdit(chatID, messageID, "admin_dates", RentDatesProps{Days: days})
}

func (h *Handler) pickAdminDate(telegramID, chatID int64, messageID int, raw string) {
	day, err := time.ParseInLocation("2006-01-02", raw, h.loc)
	if err != nil {
		h.send(chatID, "⚠️ Неверная дата. Выберите день заново.", nil)
		return
	}
	sess := h.ensureSession(telegramID, chatID)
	sess.Date = day
	sess.Hours = 0
	sess.StartsAt = time.Time{}
	h.sessions.put(telegramID, sess)
	h.showAdminHours(telegramID, chatID, messageID)
}

func (h *Handler) showAdminHours(telegramID int64, chatID int64, messageID int) {
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() {
		h.showAdminDates(chatID, messageID)
		return
	}

	hours := []CallbackBtn{
		{Callback: "adminrent:hours:1", Label: "1 ч"},
		{Callback: "adminrent:hours:2", Label: "2 ч"},
		{Callback: "adminrent:hours:3", Label: "3 ч"},
		{Callback: "adminrent:hours:4", Label: "4 ч"},
		{Callback: "adminrent:hours:5", Label: "5 ч"},
		{Callback: "adminrent:hours:6", Label: "6 ч"},
		{Callback: "adminrent:hours:7", Label: "7 ч"},
		{Callback: "adminrent:hours:8", Label: "8 ч"},
	}
	h.renderAndEdit(chatID, messageID, "admin_hours", RentHoursProps{
		DateLabel: formatDate(sess.Date),
		Hours:     hours,
	})
}

func (h *Handler) pickAdminHours(ctx context.Context, telegramID, chatID int64, messageID int, raw string) {
	hours, err := strconv.Atoi(raw)
	if err != nil || hours < 1 || hours > 8 {
		h.send(chatID, "⚠️ Выберите длительность от 1 до 8 часов.", nil)
		return
	}
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() {
		h.showAdminDates(chatID, messageID)
		return
	}
	sess.Hours = hours
	h.sessions.put(telegramID, sess)
	h.showAdminStartTimes(ctx, sess, chatID, messageID)
}

func (h *Handler) showAdminStartTimes(ctx context.Context, sess *bookingSession, chatID int64, messageID int) {
	dayStart := time.Date(sess.Date.Year(), sess.Date.Month(), sess.Date.Day(), 0, 0, 0, 0, h.loc)
	dayEnd := dayStart.Add(24 * time.Hour).Add(time.Duration(sess.Hours) * time.Hour)

	busy, err := h.rents.Occupied(ctx, dayStart, dayEnd)
	if err != nil {
		h.log.Error("ошибка загрузки занятых слотов", zap.Error(err))
		h.send(chatID, "⚠️ Не удалось загрузить свободное время.", nil)
		return
	}

	now := time.Now().In(h.loc)
	var slots []StartSlot
	for hour := studioOpenHour; hour <= studioLastStartHour; hour++ {
		start := time.Date(sess.Date.Year(), sess.Date.Month(), sess.Date.Day(), hour, 0, 0, 0, h.loc)
		end := start.Add(time.Duration(sess.Hours) * time.Hour)
		if start.Before(now) {
			continue
		}
		if overlaps(start, end, busy) {
			continue
		}
		slots = append(slots, StartSlot{
			Callback: fmt.Sprintf("adminrent:start:%d", hour),
			Label:    formatClock(start),
		})
	}

	rows := chunkSlots(slots, slotsPerRow)
	h.renderAndEdit(chatID, messageID, "admin_start", RentStartProps{
		DateLabel:  formatDate(sess.Date),
		HoursLabel: hoursLabel(sess.Hours),
		IsEmpty:    len(slots) == 0,
		Rows:       rows,
	})
}

func (h *Handler) pickAdminStart(ctx context.Context, client *domain.Client, chatID int64, messageID int, raw string) {
	hour, err := strconv.Atoi(raw)
	if err != nil || hour < studioOpenHour || hour > studioLastStartHour {
		h.send(chatID, "⚠️ Выберите время из списка.", nil)
		return
	}

	telegramID := client.TgUser.TelegramId
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() || sess.Hours == 0 {
		h.startAdminBooking(ctx, client, chatID, messageID)
		return
	}

	start := time.Date(sess.Date.Year(), sess.Date.Month(), sess.Date.Day(), hour, 0, 0, 0, h.loc)
	end := start.Add(time.Duration(sess.Hours) * time.Hour)

	created, err := h.rents.CreateUnpaid(ctx, &domain.Rent{
		ClientID:     client.ID,
		StartsAt:     start,
		EndsAt:       end,
		PricePerHour: 0,
		PaidDuration: sess.Hours,
		Type:         adminRentType,
		Notes:        "бронь администратора",
	})
	if err != nil {
		if errors.Is(err, domain.ErrSlotBusy) {
			h.send(chatID, "Этот слот уже занят. Выберите другое время.", nil)
			h.showAdminStartTimes(ctx, sess, chatID, messageID)
			return
		}
		h.log.Error("ошибка создания брони администратора", zap.Error(err))
		h.send(chatID, "⚠️ Не удалось забронировать слот. Попробуйте позже.", nil)
		return
	}

	// бронь администратора бесплатная — подтверждаем сразу, без чека
	if _, err := h.rents.ConfirmPaid(ctx, created.ID); err != nil {
		h.log.Error("ошибка подтверждения брони администратора", zap.Error(err))
	}

	h.sessions.clear(telegramID)

	h.renderAndEdit(chatID, messageID, "admin_booked", AdminBookedProps{
		DateLabel:  formatDate(start),
		TimeRange:  formatTimeRange(start, end),
		HoursLabel: hoursLabel(sess.Hours),
	})
}

// --- пропсы админского раздела ---

type AdminScheduleProps struct {
	IsTeamView bool
	IsEmpty    bool
	Items      []AdminRentItemProps
}

type AdminRentItemProps struct {
	DateLabel  string
	TimeRange  string
	HoursLabel string
	BookedBy   string // заполняется только в общем расписании команды
}

type AdminBookedProps struct {
	DateLabel  string
	TimeRange  string
	HoursLabel string
}
