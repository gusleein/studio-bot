package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/botview"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

type Bot struct {
	l       *logger.Logger
	Bot     *tgbotapi.BotAPI
	timeout int
	handler *Handler
}

func New(
	token string,
	debug bool,
	timeout int,
	clients service.ClientService,
	rents service.RentService,
	logger *logger.Logger,
	settings Settings,
) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("создание бота: %w", err)
	}
	bot.Debug = debug

	executablePath, err := os.Executable()
	if err != nil {
		err = fmt.Errorf("get executable path: %w", err)
		return nil, err
	}

	executableDir := filepath.Dir(executablePath)
	templatesDir := filepath.Join(executableDir, "templates")
	
	botviewRenderer := botview.New(templatesDir)

	loc, err := time.LoadLocation(settings.Location)
	if err != nil {
		loc = time.FixedZone("UTC", 0)
	}

	return &Bot{
		l:       logger,
		Bot:     bot,
		timeout: timeout,
		handler: NewHandler(
			bot,
			clients,
			rents,
			logger,
			botviewRenderer,
			settings,
			loc,
		),
	}, nil
}

func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.timeout

	updates := b.Bot.GetUpdatesChan(u)

	b.l.Info("бот начал получать обновления")
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			go b.handler.HandleUpdate(update)
		}
	}
}
