package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourstudio/studio-bot/pkg/logger"
	"time"
)

func New() {

}

func Run(ctx context.Context,
	timeout time.Duration,
	handler *Handler,
	log logger.Logger,
) {
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
}
