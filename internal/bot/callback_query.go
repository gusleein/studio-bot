package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (h *Handler) handleCallbackQuery(_ context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCallback(cq.ID)

	parts := strings.SplitN(cq.Data, ":", 2)
	if len(parts) != 2 {
		h.log.Warn("неверный формат callback_data", zap.String("data", cq.Data))
		return
	}

	h.log.Warn("неизвестная callback-команда", zap.String("cmd", parts[0]))
}
