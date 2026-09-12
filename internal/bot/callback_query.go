package bot

import (
	"context"
	"errors"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// handleCallbackQuery диспетчеризирует входящие callback-запросы.
// Формат callback_data: "cmd:arg", например "sched:0" или "sched_my:-1".
func (h *Handler) handleCallbackQuery(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCallback(cq.ID)

	parts := strings.SplitN(cq.Data, ":", 2)
	if len(parts) != 2 {
		h.log.Warn("неверный формат callback_data", zap.String("data", cq.Data))
		return
	}

	cmd, arg := parts[0], parts[1]

	switch cmd {
	case "sched":
		h.callbackScheduleWeek(ctx, cq, arg, false)
	case "sched_my":
		h.callbackScheduleWeek(ctx, cq, arg, true)
	default:
		h.log.Warn("неизвестная callback-команда", zap.String("cmd", cmd))
	}
}

// callbackScheduleWeek обрабатывает навигацию по неделям расписания.
func (h *Handler) callbackScheduleWeek(ctx context.Context, cq *tgbotapi.CallbackQuery, offsetStr string, isMy bool) {
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.log.Warn("неверный offset в callback", zap.String("offset", offsetStr))
		return
	}

	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	if isMy {
		student, err := h.student.GetByTelegramID(ctx, cq.From.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				h.editMessage(chatID, messageID, "👋 Вы ещё не зарегистрированы. Нажмите /start", nil)
				return
			}
			h.log.Error("ошибка получения студента в callback", zap.Error(err))
			return
		}

		schedules, err := h.schedule.GetMyWeek(ctx, student.ID, offset)
		if err != nil {
			h.log.Error("ошибка получения расписания студента в callback", zap.Error(err))
			return
		}

		h.renderAndEdit(chatID, messageID, "schedule", buildScheduleProps(schedules, offset, true))
	} else {
		schedules, err := h.schedule.GetWeek(ctx, offset)
		if err != nil {
			h.log.Error("ошибка получения расписания в callback", zap.Error(err))
			return
		}

		h.renderAndEdit(chatID, messageID, "schedule", buildScheduleProps(schedules, offset, false))
	}
}
