package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/botview"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// Handler — основной обработчик обновлений Telegram-бота.
type Handler struct {
	bot      *tgbotapi.BotAPI
	student  service.StudentService
	schedule service.ScheduleService
	sub      service.SubscriptionService
	log      *logger.Logger
	renderer *botview.Renderer
	subLinks []SubLinkProps // Tribute-ссылки для оформления подписок
}

// NewHandler создаёт новый обработчик бота.
func NewHandler(
	bot *tgbotapi.BotAPI,
	student service.StudentService,
	schedule service.ScheduleService,
	sub service.SubscriptionService,
	log *logger.Logger,
	renderer *botview.Renderer,
	subLinks []SubLinkProps,
) *Handler {
	return &Handler{
		bot:      bot,
		student:  student,
		schedule: schedule,
		sub:      sub,
		log:      log,
		renderer: renderer,
		subLinks: subLinks,
	}
}

// HandleUpdate диспетчеризирует входящее обновление Telegram.
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	ctx := context.Background()
	switch {
	case update.Message != nil:
		h.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		h.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

// renderAndSend рендерит шаблон и отправляет сообщение в чат.
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

// renderAndEdit рендерит шаблон и редактирует существующее сообщение.
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

// send отправляет текстовое HTML-сообщение с опциональной инлайн-клавиатурой.
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

// editMessage редактирует существующее сообщение (текст + разметка).
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

// answerCallback отвечает на callback query, убирая индикатор загрузки.
func (h *Handler) answerCallback(callbackID string) {
	answer := tgbotapi.NewCallback(callbackID, "")
	if _, err := h.bot.Request(answer); err != nil {
		h.log.Warn("ошибка ответа на callback",
			zap.String("callback_id", callbackID),
			zap.Error(err),
		)
	}
}
