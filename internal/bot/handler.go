package bot

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/botview"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// Settings — параметры клиентской аренды.
type Settings struct {
	AdminChatID   int64
	CardNumber    string
	PricePerHour  int
	AdminPhone    string
	AdminUsername string
	Location      string
}

// Handler — основной обработчик обновлений Telegram-бота.
type Handler struct {
	bot      *tgbotapi.BotAPI
	clients  service.ClientService
	rents    service.RentService
	log      *logger.Logger
	renderer *botview.Renderer
	sessions *sessionStore
	settings Settings
	loc      *time.Location
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	clients service.ClientService,
	rents service.RentService,
	log *logger.Logger,
	renderer *botview.Renderer,
	settings Settings,
	loc *time.Location,
) *Handler {
	return &Handler{
		bot:      bot,
		clients:  clients,
		rents:    rents,
		log:      log,
		renderer: renderer,
		sessions: newSessionStore(),
		settings: settings,
		loc:      loc,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	ctx := context.Background()
	switch {
	case update.Message != nil:
		h.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		h.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

func (h *Handler) renderAndSend(chatID int64, tmpl string, props any) {
	result, err := h.renderer.Render(tmpl, props)
	if err != nil {
		h.log.Error("ошибка рендеринга шаблона",
			zap.String("template", tmpl),
			zap.Error(err),
		)
		h.send(chatID, "⚠️ Произошла ошибка. Попробуйте позже.", nil)
		return
	}
	h.send(chatID, result.Text, result.Keyboard)
}

func (h *Handler) renderAndEdit(chatID int64, messageID int, tmpl string, props any) {
	result, err := h.renderer.Render(tmpl, props)
	if err != nil {
		h.log.Error("ошибка рендеринга шаблона",
			zap.String("template", tmpl),
			zap.Error(err),
		)
		return
	}
	h.editMessage(chatID, messageID, result.Text, result.Keyboard)
}

func (h *Handler) send(chatID int64, text string, markup interface{}) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	if markup != nil {
		msg.ReplyMarkup = markup
	}

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("ошибка отправки сообщения",
			zap.Int64("chat_id", chatID),
			zap.Error(err),
		)
	}
}

func (h *Handler) editMessage(chatID int64, messageID int, text string, markup interface{}) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "HTML"

	if markup != nil {
		if kb, ok := markup.(*tgbotapi.InlineKeyboardMarkup); ok && kb != nil {
			edit.ReplyMarkup = kb
		}
	}

	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("ошибка редактирования сообщения",
			zap.Int64("chat_id", chatID),
			zap.Int("message_id", messageID),
			zap.Error(err),
		)
	}
}

func (h *Handler) answerCallback(callbackID string) {
	answer := tgbotapi.NewCallback(callbackID, "")
	if _, err := h.bot.Request(answer); err != nil {
		h.log.Warn("ошибка ответа на callback",
			zap.String("callback_id", callbackID),
			zap.Error(err),
		)
	}
}
