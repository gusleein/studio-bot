package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/service"
)

// handleMessage — диспетчер входящих сообщений.
func (h *Handler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	if msg.Contact != nil {
		h.handleContact(ctx, msg)
		return
	}

	if !msg.IsCommand() {
		return
	}

	switch msg.Command() {
	case "start":
		h.handleStart(ctx, msg)
	default:
		h.send(msg.Chat.ID, "❓ Неизвестная команда. Попробуйте /start", nil)
	}
}

// handleStart сохраняет telegram user + client и при необходимости просит номер телефона.
func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	from := msg.From

	client, err := h.clients.GetOrCreate(ctx, service.ClientUpsert{
		TelegramID: from.ID,
		Username:   from.UserName,
		FirstName:  from.FirstName,
		LastName:   from.LastName,
	})
	if err != nil {
		h.log.Error("ошибка GetOrCreate в /start", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return
	}

	if client.TgUser.Phone == "" {
		h.requestPhone(msg.Chat.ID)
		return
	}

	h.sendWelcome(msg.Chat.ID, client.TgUser.FirstName, nil)
}

// handleContact принимает контакт из reply-клавиатуры и сохраняет телефон.
func (h *Handler) handleContact(ctx context.Context, msg *tgbotapi.Message) {
	from := msg.From
	contact := msg.Contact

	if contact.UserID != from.ID {
		h.send(msg.Chat.ID, "⚠️ Нужен ваш номер телефона. Нажмите кнопку ниже.", phoneKeyboard())
		return
	}
	if contact.PhoneNumber == "" {
		h.requestPhone(msg.Chat.ID)
		return
	}

	_, err := h.clients.GetOrCreate(ctx, service.ClientUpsert{
		TelegramID: from.ID,
		Username:   from.UserName,
		FirstName:  from.FirstName,
		LastName:   from.LastName,
	})
	if err != nil {
		h.log.Error("ошибка GetOrCreate при сохранении телефона", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Произошла ошибка. Попробуйте /start ещё раз.", nil)
		return
	}

	if err := h.clients.SavePhone(ctx, from.ID, contact.PhoneNumber); err != nil {
		h.log.Error("ошибка сохранения телефона", zap.Error(err))
		h.send(msg.Chat.ID, "⚠️ Не удалось сохранить номер. Попробуйте ещё раз.", phoneKeyboard())
		return
	}

	h.sendWelcome(msg.Chat.ID, from.FirstName, tgbotapi.NewRemoveKeyboard(true))
}

func (h *Handler) requestPhone(chatID int64) {
	h.send(chatID, "📱 Чтобы продолжить, поделитесь номером телефона — нажмите кнопку ниже.", phoneKeyboard())
}

func (h *Handler) sendWelcome(chatID int64, firstName string, markup interface{}) {
	result, err := h.renderer.Render("welcome", WelcomeProps{FirstName: firstName})
	if err != nil {
		h.log.Error("ошибка рендеринга welcome", zap.Error(err))
		h.send(chatID, "⚠️ Произошла ошибка. Попробуйте позже.", markup)
		return
	}
	h.send(chatID, result.Text, markup)
}
