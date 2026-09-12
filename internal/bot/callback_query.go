package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (h *Handler) handleCallbackQuery(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCallback(cq.ID)

	if cq.Message == nil {
		h.log.Warn("callback без сообщения", zap.String("data", cq.Data))
		return
	}

	parts := strings.SplitN(cq.Data, ":", 2)
	if len(parts) != 2 {
		h.log.Warn("неверный формат callback_data", zap.String("data", cq.Data))
		return
	}

	switch parts[0] {
	case "rent":
		h.handleRentCallback(ctx, cq, parts[1])
	case "admin":
		h.handleAdminCallback(ctx, cq, parts[1])
	default:
		h.log.Warn("неизвестная callback-команда", zap.String("cmd", parts[0]))
	}
}
