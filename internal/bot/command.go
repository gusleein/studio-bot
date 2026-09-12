package bot

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// handleMessage — диспетчер входящих сообщений.
func (h *Handler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return
	}

	switch msg.Command() {
	case "start":
		h.handleStart(ctx, msg)
	case "schedule":
		h.handleSchedule(ctx, msg)
	case "my":
		h.handleMy(ctx, msg)
	case "subscription":
		h.handleSubscription(ctx, msg)
	default:
		h.send(msg.Chat.ID, "❓ Неизвестная команда. Попробуйте /start", nil)
	}
}

// handleStart регистрирует студента и отправляет приветственное сообщение.
func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	from := msg.From

	student, err := h.student.GetOrCreate(ctx, from.ID, from.FirstName, from.LastName, from.UserName)
	if err != nil {
		h.log.Error("ошибка GetOrCreate в /start", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return
	}

	h.renderAndSend(msg.Chat.ID, "welcome", WelcomeProps{
		FirstName: student.FirstName,
	})
}

// handleSchedule отправляет общее расписание на текущую неделю.
func (h *Handler) handleSchedule(ctx context.Context, msg *tgbotapi.Message) {
	schedules, err := h.schedule.GetWeek(ctx, 0)
	if err != nil {
		h.log.Error("ошибка получения расписания", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Не удалось загрузить расписание. Попробуйте позже.", nil)
		return
	}

	h.renderAndSend(msg.Chat.ID, "schedule", buildScheduleProps(schedules, 0, false))
}

// handleMy отправляет персональное расписание студента.
func (h *Handler) handleMy(ctx context.Context, msg *tgbotapi.Message) {
	student, err := h.student.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.send(msg.Chat.ID, "👋 Вы ещё не зарегистрированы. Нажмите /start", nil)
			return
		}
		h.log.Error("ошибка получения студента в /my", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return
	}

	schedules, err := h.schedule.GetMyWeek(ctx, student.ID, 0)
	if err != nil {
		h.log.Error("ошибка получения расписания студента", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Не удалось загрузить расписание. Попробуйте позже.", nil)
		return
	}

	h.renderAndSend(msg.Chat.ID, "schedule", buildScheduleProps(schedules, 0, true))
}

// handleSubscription отправляет информацию об активных подписках.
func (h *Handler) handleSubscription(ctx context.Context, msg *tgbotapi.Message) {
	student, err := h.student.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.send(msg.Chat.ID, "👋 Вы ещё не зарегистрированы. Нажмите /start", nil)
			return
		}
		h.log.Error("ошибка получения студента в /subscription", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return
	}

	subs, err := h.sub.GetActive(ctx, student.ID)
	if err != nil {
		h.log.Error("ошибка получения подписок", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Не удалось загрузить подписки. Попробуйте позже.", nil)
		return
	}

	h.renderAndSend(msg.Chat.ID, "subscription", buildSubscriptionProps(subs, h.subLinks))
}
