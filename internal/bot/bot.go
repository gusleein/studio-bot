package bot

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

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
	logger *logger.Logger,
) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("создание бота: %w", err)
	}
	bot.Debug = debug

	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	templatesDir := filepath.Join(projectRoot, "internal/bot/templates")
	botviewRenderer := botview.New(templatesDir)

	return &Bot{
		l:       logger,
		Bot:     bot,
		timeout: timeout,
		handler: NewHandler(
			bot,
			clients,
			logger,
			botviewRenderer,
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
