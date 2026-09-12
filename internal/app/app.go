package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/yourstudio/studio-bot/internal/config"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source для migrate
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver для sqlx
	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/pkg/logger"
)

// App — корневой объект приложения.
type App struct {
	cfg             *config.Config
	db              *sqlx.DB
	httpServer      *http.Server
	bot             *tgbotapi.BotAPI
	log             *logger.Logger
	serviceProvider *serviceProvider
}

// New создаёт и инициализирует приложение.
func New(ctx context.Context, log *logger.Logger) (*App, error) {
	cfg := config.Load()

	a := &App{
		cfg: &cfg,
		log: log,
	}

	if err := a.initDeps(ctx); err != nil {
		return nil, fmt.Errorf("инициализация зависимостей: %w", err)
	}

	return a, nil
}

func Run(ctx context.Context, log *logger.Logger) {
	app, err := New(ctx, log)
	if err != nil {
		log.Fatal("app init error", zap.Error(err))
	}
	if err = app.Run(ctx); err != nil {
		log.Fatal("приложение завершилось с ошибкой", zap.Error(err))
	}
}

// initDeps инициализирует все зависимости приложения.
func (a *App) initDeps(ctx context.Context) error {
	if err := a.initDB(); err != nil {
		return fmt.Errorf("инициализация БД: %w", err)
	}

	if err := a.runMigrations(); err != nil {
		return fmt.Errorf("выполнение миграций: %w", err)
	}

	if err := a.initBot(); err != nil {
		return fmt.Errorf("инициализация бота: %w", err)
	}

	a.serviceProvider = newServiceProvider(a.cfg, a.db, a.bot, a.log)

	return nil
}

// initDB создаёт подключение к PostgreSQL.
func (a *App) initDB() error {
	db, err := sqlx.Connect("postgres", a.cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("подключение к PostgreSQL: %w", err)
	}

	db.SetMaxOpenConns(a.cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(a.cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(a.cfg.DB.ConnMaxLifetime)

	a.db = db
	a.log.Info("подключение к БД установлено",
		zap.String("host", a.cfg.DB.Host),
		zap.Int("port", a.cfg.DB.Port),
	)
	return nil
}

// runMigrations применяет все ожидающие миграции.
func (a *App) runMigrations() error {
	m, err := migrate.New("file://migrations", a.cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("инициализация migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("применение миграций: %w", err)
	}

	a.log.Info("миграции применены")
	return nil
}

// initBot создаёт клиента Telegram Bot API.
func (a *App) initBot() error {
	bot, err := tgbotapi.NewBotAPI(a.cfg.Bot.Token)
	if err != nil {
		return fmt.Errorf("создание бота: %w", err)
	}
	bot.Debug = a.cfg.Bot.Debug

	a.bot = bot
	a.log.Info("бот инициализирован", zap.String("username", bot.Self.UserName))
	return nil
}

// Run запускает все компоненты приложения и ждёт завершения контекста.
func (a *App) Run(ctx context.Context) error {
	// Запуск Telegram Bot polling
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = a.cfg.Bot.Timeout

		updates := a.bot.GetUpdatesChan(u)
		handler := a.serviceProvider.BotHandler()

		a.log.Info("бот начал получать обновления")
		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				go handler.HandleUpdate(update)
			}
		}
	}()

	// Ждём сигнала завершения
	<-ctx.Done()
	a.log.Info("начало graceful shutdown")
	return a.GracefulShutdown()
}

// GracefulShutdown корректно завершает все компоненты приложения.
func (a *App) GracefulShutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.log.Error("ошибка завершения HTTP сервера", zap.Error(err))
	}

	a.bot.StopReceivingUpdates()

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("закрытие соединения с БД: %w", err)
	}

	a.log.Info("приложение завершено")
	return nil
}
