package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/yourstudio/studio-bot/internal/app"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic("не удалось инициализировать логгер: " + err.Error())
	}
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go app.Run(ctx, log)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh
	log.Info("получен сигнал завершения", zap.String("signal", sig.String()))
	cancel()

}
