package api

import (
	"errors"
	"github.com/yourstudio/studio-bot/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/logger"
	"github.com/yourstudio/studio-bot/pkg/tribute"
)

// Handler — обработчик вебхуков от Tribute.
type Handler struct {
	cfg         config.TributeConfig
	studentSvc  service.StudentService
	subSvc      service.SubscriptionService
	paymentRepo repository.PaymentRepository
	bot         *tgbotapi.BotAPI
	log         *logger.Logger
}

// New создаёт новый обработчик вебхуков Tribute.
func New(
	cfg config.TributeConfig,
	studentSvc service.StudentService,
	subSvc service.SubscriptionService,
	paymentRepo repository.PaymentRepository,
	bot *tgbotapi.BotAPI,
	log *logger.Logger,
) *Handler {
	return &Handler{
		cfg:         cfg,
		studentSvc:  studentSvc,
		subSvc:      subSvc,
		paymentRepo: paymentRepo,
		bot:         bot,
		log:         log,
	}
}

// RegisterRoutes регистрирует маршрут вебхука.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tribute", h.Handle)
}

// Handle обрабатывает POST /webhook/tribute.
// Порядок обработки:
//  1. Читаем сырое тело до Bind.
//  2. Верифицируем HMAC-подпись → 401.
//  3. Парсим payload → 400.
//  4. Проверяем payload.Name == "new_digital_product".
//  5. Определяем SubscriptionType по ProductID.
//  6. Идемпотентность: ExistsByTributeData → 200 если дубль.
//  7. Получаем студента по TelegramUserID → 404 если нет.
//  8. Активируем подписку.
//  9. Отправляем подтверждение в Telegram.
//  10. JSON 200 {"ok": true}.
func (h *Handler) Handle(c *gin.Context) {
	// 1. Читаем тело до любых попыток bind
	body, err := c.GetRawData()
	if err != nil {
		h.log.Error("ошибка чтения тела вебхука", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "ошибка чтения запроса"})
		return
	}

	// 2. Верификация подписи
	signature := c.GetHeader("X-Tribute-Signature")
	if err := tribute.VerifySignature(body, signature, h.cfg.APIKey); err != nil {
		h.log.Warn("неверная подпись вебхука Tribute", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверная подпись"})
		return
	}

	// 3. Парсим payload
	payload, err := tribute.ParsePayload(body)
	if err != nil {
		h.log.Warn("ошибка парсинга вебхука Tribute", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат данных"})
		return
	}

	// 4. Обрабатываем только события нового цифрового продукта
	if payload.Name != "new_digital_product" {
		h.log.Info("игнорируем событие Tribute", zap.String("name", payload.Name))
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 5. Определяем тип подписки по ProductID
	subType, err := h.resolveSubscriptionType(payload.Payload.ProductID)
	if err != nil {
		h.log.Warn("неизвестный ProductID в вебхуке",
			zap.Int("product_id", payload.Payload.ProductID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неизвестный продукт"})
		return
	}

	ctx := c.Request.Context()

	// 6. Идемпотентность — проверяем, не обрабатывали ли уже этот платёж
	// ProductPayload.UserID имеет тип int (Tribute user id)
	tributeUserID := int64(payload.Payload.UserID)
	exists, err := h.paymentRepo.ExistsByTributeData(ctx,
		tributeUserID,
		payload.Payload.ProductID,
		payload.Payload.Amount,
	)
	if err != nil {
		h.log.Error("ошибка проверки платежа", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}
	if exists {
		h.log.Info("платёж уже обработан, пропускаем",
			zap.Int64("tribute_user_id", tributeUserID),
		)
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	// 7. Получаем студента по Telegram ID
	telegramID := payload.Payload.TelegramUserID
	student, err := h.studentSvc.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.log.Warn("студент не найден в вебхуке Tribute",
				zap.Int64("telegram_id", telegramID),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": "студент не найден"})
			return
		}
		h.log.Error("ошибка поиска студента в вебхуке", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка сервера"})
		return
	}

	// 8. Активируем подписку
	_, err = h.subSvc.Activate(ctx, service.ActivateRequest{
		StudentID:        student.ID,
		Type:             subType,
		TributeProductID: payload.Payload.ProductID,
		TributeUserID:    tributeUserID,
		Amount:           payload.Payload.Amount,
		Currency:         payload.Payload.Currency,
		DurationDays:     30,
	})
	if err != nil {
		h.log.Error("ошибка активации подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка активации подписки"})
		return
	}

	// 9. Отправляем подтверждение студенту в Telegram
	h.sendConfirmation(telegramID, subType)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// resolveSubscriptionType определяет тип подписки по ProductID из конфига Tribute.
func (h *Handler) resolveSubscriptionType(productID int) (domain.SubscriptionType, error) {
	switch productID {
	case h.cfg.ProductIDSamplePacks:
		return domain.SubscriptionTypeSamplePacks, nil
	case h.cfg.ProductIDPresets:
		return domain.SubscriptionTypePresets, nil
	case h.cfg.ProductIDBundle:
		return domain.SubscriptionTypeBundle, nil
	default:
		return "", errors.New("неизвестный product_id")
	}
}

// sendConfirmation отправляет студенту сообщение об активации подписки.
func (h *Handler) sendConfirmation(telegramID int64, subType domain.SubscriptionType) {
	typeNames := map[domain.SubscriptionType]string{
		domain.SubscriptionTypeSamplePacks: "🎵 Sample Packs",
		domain.SubscriptionTypePresets:     "🎛 Presets",
		domain.SubscriptionTypeBundle:      "📦 Bundle",
	}
	name := typeNames[subType]

	text := "✅ <b>Подписка активирована!</b>\n\n" +
		"Тип: " + name + "\n" +
		"Доступ открыт на 30 дней.\n\n" +
		"Спасибо за поддержку! 🎶"

	msg := tgbotapi.NewMessage(telegramID, text)
	msg.ParseMode = "HTML"

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Warn("не удалось отправить подтверждение подписки",
			zap.Int64("telegram_id", telegramID),
			zap.Error(err),
		)
	}
}
