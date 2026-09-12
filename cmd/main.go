package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourstudio/studio-bot/internal/app"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

func main() {
	log, err := logger.New()
	if err != nil {
		panic("не удалось инициализировать логгер: " + err.Error())
	}
	defer log.Sync()

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())

	// Инициализируем приложение
	a := app.NewApp(log)

	// Запускаем приложение
	go a.Run(ctx)

	// Ожидаем сигнал завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	a.GracefulShutdown(cancel)
}
