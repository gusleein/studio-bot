package app

import (
	"github.com/yourstudio/studio-bot/internal/api"
	"github.com/yourstudio/studio-bot/internal/config"
	"github.com/yourstudio/studio-bot/internal/repository/payment"
	"github.com/yourstudio/studio-bot/pkg/auth"
	"path/filepath"
	"runtime"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/bot"
	"github.com/yourstudio/studio-bot/internal/repository"

	"github.com/yourstudio/studio-bot/pkg/botview"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// serviceProvider — контейнер зависимостей с ленивой инициализацией через sync.Once.
type serviceProvider struct {
	cfg *config.Config
	db  *sqlx.DB
	bot *tgbotapi.BotAPI
	log *logger.Logger

	paymentRepoOnce sync.Once
	paymentRepo     repository.PaymentRepository

	// JWT
	jwtManagerOnce sync.Once
	jwtManager     *auth.Manager

	// HTTP обработчики
	webhookHandlerOnce sync.Once
	webhookHandler     *api.Handler

	// Botview renderer
	botviewRendererOnce sync.Once
	botviewRenderer     *botview.Renderer

	// Bot обработчик
	botHandlerOnce sync.Once
	botHandler     *bot.Handler
}

func newServiceProvider(cfg *config.Config, db *sqlx.DB, botAPI *tgbotapi.BotAPI, log *logger.Logger) *serviceProvider {
	return &serviceProvider{cfg: cfg, db: db, bot: botAPI, log: log}
}

// ──────────────────── Репозитории ────────────────────

func (sp *serviceProvider) PaymentRepo() repository.PaymentRepository {
	sp.paymentRepoOnce.Do(func() {
		sp.paymentRepo = payment.New(sp.db)
	})
	return sp.paymentRepo
}

// ──────────────────── HTTP обработчики ────────────────────

func (sp *serviceProvider) StudentHandler() *apiStudent.Handler {
	sp.studentHandlerOnce.Do(func() {
		sp.studentHandler = apiStudent.New(sp.StudentSvc(), sp.JWTManager(), sp.log)
	})
	return sp.studentHandler
}

func (sp *serviceProvider) ScheduleHandler() *apiSchedule.Handler {
	sp.scheduleHandlerOnce.Do(func() {
		sp.scheduleHandler = apiSchedule.New(sp.ScheduleSvc(), sp.JWTManager(), sp.log)
	})
	return sp.scheduleHandler
}

func (sp *serviceProvider) SubHandler() *apiSubscription.Handler {
	sp.subHandlerOnce.Do(func() {
		sp.subHandler = apiSubscription.New(sp.StudentSvc(), sp.SubscriptionSvc(), sp.JWTManager(), sp.log)
	})
	return sp.subHandler
}

func (sp *serviceProvider) WebhookHandler() *api.Handler {
	sp.webhookHandlerOnce.Do(func() {
		sp.webhookHandler = api.New(
			sp.cfg.Tribute,
			sp.StudentSvc(),
			sp.SubscriptionSvc(),
			sp.PaymentRepo(),
			sp.bot,
			sp.log,
		)
	})
	return sp.webhookHandler
}

func (sp *serviceProvider) BotviewRenderer() *botview.Renderer {
	sp.botviewRendererOnce.Do(func() {
		// Определяем путь к директории шаблонов относительно корня проекта
		_, filename, _, _ := runtime.Caller(0)
		projectRoot := filepath.Join(filepath.Dir(filename), "../..")
		templatesDir := filepath.Join(projectRoot, "internal/bot/templates")
		sp.botviewRenderer = botview.New(templatesDir)
	})
	return sp.botviewRenderer
}

func (sp *serviceProvider) BotHandler() *bot.Handler {
	sp.botHandlerOnce.Do(func() {
		subLinks := []bot.SubLinkProps{}

		sp.botHandler = bot.NewHandler(
			sp.bot,
			sp.StudentSvc(),
			sp.ScheduleSvc(),
			sp.SubscriptionSvc(),
			sp.log,
			sp.BotviewRenderer(),
			subLinks,
		)
	})
	return sp.botHandler
}
