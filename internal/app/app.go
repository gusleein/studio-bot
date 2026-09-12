package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source для migrate
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver для sqlx
	"github.com/yourstudio/studio-bot/internal/bot"
	"github.com/yourstudio/studio-bot/internal/config"
	"go.uber.org/zap"
	"net/http"

	"github.com/yourstudio/studio-bot/pkg/logger"
)

// App — корневой объект приложения.
type App struct {
	cfg             *config.Config
	db              *sqlx.DB
	httpServer      *http.Server
	bot             *bot.Bot
	log             *logger.Logger
	serviceProvider *serviceProvider
}

func NewApp(log *logger.Logger) *App {
	a := &App{
		log: log,
	}
	a.initDeps(context.Background())
	return a
}

func (a *App) Run(ctx context.Context) {
	a.log.Info("starting application")

	go a.bot.Run(ctx)
}

func (a *App) initDeps(ctx context.Context) {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initDB,
		a.runMigrations,
		a.initServiceProvider,
		a.initBot,
	}

	for _, f := range inits {
		if err := f(ctx); err != nil {
			a.log.Fatal(fmt.Sprintf("failed to init deps: %v", err))
		}
	}
}

func (a *App) initConfig(_ context.Context) error {
	a.log.Info("loading config")
	cfg := config.Load()
	a.cfg = &cfg
	return nil
}

// initDB создаёт подключение к PostgreSQL.
func (a *App) initDB(_ context.Context) error {
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
func (a *App) runMigrations(_ context.Context) error {
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
func (a *App) initBot(_ context.Context) error {
	var err error
	a.bot, err = bot.New(a.cfg.Bot.Token,
		a.cfg.Bot.Debug,
		a.cfg.Bot.Timeout,
		a.serviceProvider.ClientsService(),
		a.log)
	return err
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.log.Info("initializing service provider")
	a.serviceProvider = newServiceProvider(
		a.cfg,
		a.db,
		a.log)
	return nil
}

// GracefulShutdown корректно завершает все компоненты приложения.
func (a *App) GracefulShutdown(cancel context.CancelFunc) error {
	//shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	//if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
	//	a.log.Error("ошибка завершения HTTP сервера", zap.Error(err))
	//}

	a.bot.Bot.StopReceivingUpdates()

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("закрытие соединения с БД: %w", err)
	}

	a.log.Info("приложение завершено")
	return nil
}
