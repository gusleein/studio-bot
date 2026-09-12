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
	"github.com/yourstudio/studio-bot/internal/service"
)

const (
	rentTypeDJ          = "dj"
	studioOpenHour      = 10
	studioLastStartHour = 22
	rentDaysAhead       = 7
	slotsPerRow         = 3
)

func (h *Handler) handleRent(ctx context.Context, msg *tgbotapi.Message) {
	client, ok := h.requireClient(ctx, msg.From, msg.Chat.ID)
	if !ok {
		return
	}
	_ = client
	h.showRentMenu(msg.Chat.ID, 0)
}

func (h *Handler) handleRentCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, payload string) {
	client, ok := h.requireClient(ctx, cq.From, cq.Message.Chat.ID)
	if !ok {
		return
	}

	chatID := cq.Message.Chat.ID
	msgID := cq.Message.MessageID

	switch {
	case payload == "menu":
		h.showRentMenu(chatID, msgID)
	case payload == "book":
		h.startBooking(ctx, client, chatID, msgID)
	case payload == "next":
		h.showUpcoming(ctx, client, chatID, msgID)
	case payload == "history":
		h.showHistory(ctx, client, chatID, msgID)
	case payload == "hours_back":
		h.showHours(cq.From.ID, chatID, msgID)
	case strings.HasPrefix(payload, "date:"):
		h.pickDate(cq.From.ID, chatID, msgID, strings.TrimPrefix(payload, "date:"))
	case strings.HasPrefix(payload, "hours:"):
		h.pickHours(ctx, cq.From.ID, chatID, msgID, strings.TrimPrefix(payload, "hours:"))
	case strings.HasPrefix(payload, "start:"):
		h.pickStart(ctx, client, cq.From.ID, chatID, msgID, strings.TrimPrefix(payload, "start:"))
	default:
		h.log.Warn("неизвестный rent callback", zap.String("payload", payload))
	}
}

func (h *Handler) handleAdminCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, payload string) {
	if cq.Message == nil || cq.Message.Chat.ID != h.settings.AdminChatID {
		return
	}
	if !strings.HasPrefix(payload, "confirm:") &&
		!strings.HasPrefix(payload, "cancel:") {
		return
	}
	if strings.HasPrefix(payload, "confirm:") {
		id, err := uuid.Parse(strings.TrimPrefix(payload, "confirm:"))
		if err != nil {
			h.log.Warn("неверный id аренды в callback", zap.Error(err))
			return
		}

		rent, err := h.rents.ConfirmPaid(ctx, id)
		if err != nil {
			h.log.Error("ошибка подтверждения аренды", zap.Error(err))
			h.send(cq.Message.Chat.ID, "⚠️ Не удалось подтвердить аренду.", nil)
			return
		}

		client, err := h.clients.GetByID(ctx, rent.ClientID)
		if err != nil {
			h.log.Error("клиент для подтверждённой аренды не найден", zap.Error(err))
			client = &domain.Client{}
		} else {
			h.renderAndSend(client.TgUser.TelegramId, "rent_confirmed", toConfirmedProps(rent, h.loc))
		}

		h.renderAndEdit(cq.Message.Chat.ID, cq.Message.MessageID, "rent_admin_done", toAdminDoneProps(rent, client, h.loc))
	}

	//if strings.HasPrefix(payload, "cancel:") {
	//	id, err := uuid.Parse(strings.TrimPrefix(payload, "cancel:"))
	//	if err != nil {
	//		h.log.Warn("неверный id аренды в callback", zap.Error(err))
	//		return
	//	}
	//
	//	err := h.rents.CancelUnpaid(ctx, id)
	//	if err != nil {
	//		h.log.Error("ошибка подтверждения аренды", zap.Error(err))
	//		h.send(cq.Message.Chat.ID, "⚠️ Не удалось подтвердить аренду.", nil)
	//		return
	//	}
	//
	//	rent, err := h.rents.GetByID(ctx, id)
	//	if err != nil {
	//		h.log.Error("")
	//	}
	//
	//	client, err := h.clients.GetByID(ctx, rent.ClientID)
	//	if err != nil {
	//		h.log.Error("клиент для подтверждённой аренды не найден", zap.Error(err))
	//		client = &domain.Client{}
	//	} else {
	//		h.renderAndSend(client.TgUser.TelegramId, "rent_confirmed", toConfirmedProps(rent, h.loc))
	//	}
	//
	//	h.renderAndEdit(cq.Message.Chat.ID, cq.Message.MessageID, "rent_admin_done", toAdminDoneProps(rent, client, h.loc))
	//}

}

func (h *Handler) handleReceipt(ctx context.Context, msg *tgbotapi.Message) {
	sess := h.sessions.get(msg.From.ID)
	if sess == nil || !sess.AwaitingReceipt || sess.RentID == uuid.Nil {
		return
	}

	rent, err := h.rents.GetByID(ctx, sess.RentID)
	if err != nil {
		h.log.Error("аренда для чека не найдена", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Бронь не найдена. Начните заново через меню.", nil)
		h.sessions.clear(msg.From.ID)
		return
	}

	if _, err := h.bot.Send(tgbotapi.NewCopyMessage(h.settings.AdminChatID, msg.Chat.ID, msg.MessageID)); err != nil {
		h.log.Error("не удалось переслать чек админу", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Не удалось отправить чек. Попробуйте ещё раз.", nil)
		return
	}

	client, err := h.clients.GetByID(ctx, sess.ClientID)
	if err != nil {
		h.log.Error("клиент для чека не найден", zap.Error(err))
		client = &domain.Client{TgUser: domain.TelegramUser{FirstName: msg.From.FirstName, Username: msg.From.UserName}}
	}

	h.renderAndSend(h.settings.AdminChatID, "rent_admin_confirm", toAdminConfirmProps(rent, client, h.loc))
	h.renderAndSend(msg.Chat.ID, "rent_receipt_sent", RentReceiptSentProps{
		DateLabel: formatDate(rent.StartsAt.In(h.loc)),
		TimeRange: formatTimeRange(rent.StartsAt.In(h.loc), rent.EndsAt.In(h.loc)),
		Amount:    formatAmount(rent.PricePerHour * rent.PaidDuration),
	})
	sess.AwaitingReceipt = false
	h.sessions.put(msg.From.ID, sess)
}

func (h *Handler) showRentMenu(chatID int64, messageID int) {
	if messageID > 0 {
		h.renderAndEdit(chatID, messageID, "rent_menu", RentMenuProps{})
		return
	}
	h.renderAndSend(chatID, "rent_menu", RentMenuProps{})
}

func (h *Handler) startBooking(ctx context.Context, client *domain.Client, chatID int64, messageID int) {
	if sess := h.sessions.get(client.TgUser.TelegramId); sess != nil && sess.RentID != uuid.Nil {
		_ = h.rents.CancelUnpaid(ctx, sess.RentID)
	}

	h.sessions.put(client.TgUser.TelegramId, &bookingSession{
		ClientID: client.ID,
		ChatID:   chatID,
	})
	h.showDates(chatID, messageID)
}

func (h *Handler) showDates(chatID int64, messageID int) {
	now := time.Now().In(h.loc)
	days := make([]CallbackBtn, 0, rentDaysAhead)
	for i := 0; i < rentDaysAhead; i++ {
		d := now.AddDate(0, 0, i)
		label := formatDate(d)
		if i == 0 {
			label = "Сегодня, " + strconv.Itoa(d.Day()) + " " + ruMonths[d.Month()]
		}
		days = append(days, CallbackBtn{
			Callback: "rent:date:" + d.Format("2006-01-02"),
			Label:    label,
		})
	}
	h.renderAndEdit(chatID, messageID, "rent_dates", RentDatesProps{Days: days})
}

func (h *Handler) pickDate(telegramID, chatID int64, messageID int, raw string) {
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
	h.showHours(telegramID, chatID, messageID)
}

func (h *Handler) showHours(telegramID int64, chatID int64, messageID int) {
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() {
		h.showDates(chatID, messageID)
		return
	}

	hours := []CallbackBtn{
		{Callback: "rent:hours:1", Label: "1 ч"},
		{Callback: "rent:hours:2", Label: "2 ч"},
		{Callback: "rent:hours:3", Label: "3 ч"},
		{Callback: "rent:hours:4", Label: "4 ч"},
		{Callback: "rent:hours:5", Label: "5 ч"},
		{Callback: "rent:hours:6", Label: "6 ч"},
		{Callback: "rent:hours:7", Label: "7 ч"},
		{Callback: "rent:hours:8", Label: "8 ч"},
	}
	h.renderAndEdit(chatID, messageID, "rent_hours", RentHoursProps{
		DateLabel: formatDate(sess.Date),
		Hours:     hours,
	})
}

func (h *Handler) pickHours(ctx context.Context, telegramID, chatID int64, messageID int, raw string) {
	hours, err := strconv.Atoi(raw)
	if err != nil || hours < 1 || hours > 4 {
		h.send(chatID, "⚠️ Выберите длительность 1–4 часа.", nil)
		return
	}
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() {
		h.showDates(chatID, messageID)
		return
	}
	sess.Hours = hours
	h.sessions.put(telegramID, sess)
	h.showStartTimes(ctx, sess, chatID, messageID)
}

func (h *Handler) showStartTimes(ctx context.Context, sess *bookingSession, chatID int64, messageID int) {
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
			Callback: fmt.Sprintf("rent:start:%d", hour),
			Label:    formatClock(start),
		})
	}

	rows := chunkSlots(slots, slotsPerRow)
	h.renderAndEdit(chatID, messageID, "rent_start", RentStartProps{
		DateLabel:  formatDate(sess.Date),
		HoursLabel: hoursLabel(sess.Hours),
		IsEmpty:    len(slots) == 0,
		Rows:       rows,
	})
}

func (h *Handler) pickStart(ctx context.Context, client *domain.Client, telegramID, chatID int64, messageID int, raw string) {
	hour, err := strconv.Atoi(raw)
	if err != nil || hour < studioOpenHour || hour > studioLastStartHour {
		h.send(chatID, "⚠️ Выберите время из списка.", nil)
		return
	}
	sess := h.sessions.get(telegramID)
	if sess == nil || sess.Date.IsZero() || sess.Hours == 0 {
		h.startBooking(ctx, client, chatID, messageID)
		return
	}

	if sess.RentID != uuid.Nil {
		_ = h.rents.CancelUnpaid(ctx, sess.RentID)
		sess.RentID = uuid.Nil
	}

	start := time.Date(sess.Date.Year(), sess.Date.Month(), sess.Date.Day(), hour, 0, 0, 0, h.loc)
	end := start.Add(time.Duration(sess.Hours) * time.Hour)

	created, err := h.rents.CreateUnpaid(ctx, &domain.Rent{
		ClientID:     client.ID,
		StartsAt:     start,
		EndsAt:       end,
		PricePerHour: h.settings.PricePerHour,
		PaidDuration: sess.Hours,
		Type:         rentTypeDJ,
		Notes:        "бронь из бота",
	})
	if err != nil {
		if errors.Is(err, domain.ErrSlotBusy) {
			h.send(chatID, "Этот слот уже занят. Выберите другое время.", nil)
			h.showStartTimes(ctx, sess, chatID, messageID)
			return
		}
		h.log.Error("ошибка создания аренды", zap.Error(err))
		h.send(chatID, "⚠️ Не удалось забронировать слот. Попробуйте позже.", nil)
		return
	}

	sess.StartsAt = start
	sess.RentID = created.ID
	sess.AwaitingReceipt = true
	h.sessions.put(telegramID, sess)

	h.renderAndEdit(chatID, messageID, "rent_pay", RentPayProps{
		DateLabel:  formatDate(start.In(h.loc)),
		TimeRange:  formatTimeRange(start.In(h.loc), end.In(h.loc)),
		HoursLabel: hoursLabel(sess.Hours),
		Amount:     formatAmount(h.settings.PricePerHour * sess.Hours),
		CardNumber: h.settings.CardNumber,
	})
}

func (h *Handler) showUpcoming(ctx context.Context, client *domain.Client, chatID int64, messageID int) {
	rent, err := h.rents.Upcoming(ctx, client.ID)
	if err != nil {
		h.log.Error("ошибка ближайшей аренды", zap.Error(err))
		h.send(chatID, "⚠️ Не удалось загрузить аренду.", nil)
		return
	}

	props := RentUpcomingProps{HasRent: rent != nil}
	if rent != nil {
		props.Item = toItemProps(rent, h.loc)
	}
	h.renderAndEdit(chatID, messageID, "rent_upcoming", props)
}

func (h *Handler) showHistory(ctx context.Context, client *domain.Client, chatID int64, messageID int) {
	items, err := h.rents.History(ctx, client.ID)
	if err != nil {
		h.log.Error("ошибка истории аренды", zap.Error(err))
		h.send(chatID, "⚠️ Не удалось загрузить историю.", nil)
		return
	}

	props := RentHistoryProps{IsEmpty: len(items) == 0}
	for _, r := range items {
		props.Items = append(props.Items, toItemProps(r, h.loc))
	}
	h.renderAndEdit(chatID, messageID, "rent_history", props)
}

func (h *Handler) ensureSession(telegramID, chatID int64) *bookingSession {
	if sess := h.sessions.get(telegramID); sess != nil {
		return sess
	}
	sess := &bookingSession{ChatID: chatID}
	h.sessions.put(telegramID, sess)
	return sess
}

func (h *Handler) requireClient(ctx context.Context, from *tgbotapi.User, chatID int64) (*domain.Client, bool) {
	if from == nil {
		return nil, false
	}
	client, err := h.clients.GetOrCreate(ctx, service.ClientUpsert{
		TelegramID: from.ID,
		Username:   from.UserName,
		FirstName:  from.FirstName,
		LastName:   from.LastName,
	})
	if err != nil {
		h.log.Error("ошибка GetOrCreate", zap.Error(err))
		h.send(chatID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return nil, false
	}
	if client.TgUser.Phone == "" {
		h.requestPhone(chatID)
		return nil, false
	}
	return client, true
}

func overlaps(start, end time.Time, busy []*domain.Rent) bool {
	for _, r := range busy {
		rEnd := r.EndsAt
		if rEnd.IsZero() {
			rEnd = r.StartsAt.Add(time.Hour)
		}
		if start.Before(rEnd) && end.After(r.StartsAt) {
			return true
		}
	}
	return false
}

func chunkSlots(slots []StartSlot, n int) []StartSlotRow {
	if len(slots) == 0 {
		return nil
	}
	var rows []StartSlotRow
	for i := 0; i < len(slots); i += n {
		end := i + n
		if end > len(slots) {
			end = len(slots)
		}
		rows = append(rows, StartSlotRow{Slots: slots[i:end]})
	}
	return rows
}

func toItemProps(r *domain.Rent, loc *time.Location) RentItemProps {
	start := r.StartsAt.In(loc)
	end := r.EndsAt.In(loc)
	return RentItemProps{
		DateLabel:  formatDate(start),
		TimeRange:  formatTimeRange(start, end),
		HoursLabel: hoursLabel(r.PaidDuration),
		Amount:     formatAmount(r.PricePerHour * r.PaidDuration),
		Status:     rentStatusLabel(r.IsPaid, r.IsCancelled),
	}
}

func toConfirmedProps(r *domain.Rent, loc *time.Location) RentConfirmedProps {
	start := r.StartsAt.In(loc)
	end := r.EndsAt.In(loc)
	return RentConfirmedProps{
		DateLabel:  formatDate(start),
		TimeRange:  formatTimeRange(start, end),
		HoursLabel: hoursLabel(r.PaidDuration),
	}
}

func toAdminConfirmProps(r *domain.Rent, client *domain.Client, loc *time.Location) RentAdminConfirmProps {
	start := r.StartsAt.In(loc)
	end := r.EndsAt.In(loc)
	name := strings.TrimSpace(client.TgUser.FirstName + " " + client.TgUser.LastName)
	return RentAdminConfirmProps{
		ClientName:      name,
		Username:        client.TgUser.Username,
		Phone:           client.TgUser.Phone,
		DateLabel:       formatDate(start),
		TimeRange:       formatTimeRange(start, end),
		HoursLabel:      hoursLabel(r.PaidDuration),
		Amount:          formatAmount(r.PricePerHour * r.PaidDuration),
		ConfirmCallback: "admin:confirm:" + r.ID.String(),
		CancelCallback:  "admin:cancel:" + r.ID.String(),
	}
}

func toAdminDoneProps(r *domain.Rent, client *domain.Client, loc *time.Location) RentAdminConfirmProps {
	props := toAdminConfirmProps(r, client, loc)
	if client == nil {
		props.ClientName = "клиент"
	}
	return props
}
