package app

import (
	"github.com/yourstudio/studio-bot/internal/service"
	client2 "github.com/yourstudio/studio-bot/internal/service/client"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"

	"github.com/yourstudio/studio-bot/internal/config"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/client"
	"github.com/yourstudio/studio-bot/internal/repository/payment"
	"github.com/yourstudio/studio-bot/internal/repository/product"
	"github.com/yourstudio/studio-bot/internal/repository/rent"
	"github.com/yourstudio/studio-bot/internal/repository/telegramuser"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

// serviceProvider — контейнер зависимостей с ленивой инициализацией через sync.Once.
type serviceProvider struct {
	cfg *config.Config
	db  *sqlx.DB
	bot *tgbotapi.BotAPI
	log *logger.Logger

	telegramUserRepoOnce sync.Once
	telegramUserRepo     repository.TelegramUserRepository

	clientRepoOnce sync.Once
	clientRepo     repository.ClientRepository

	rentRepoOnce sync.Once
	rentRepo     repository.RentRepository

	productRepoOnce sync.Once
	productRepo     repository.ProductRepository

	paymentRepoOnce sync.Once
	paymentRepo     repository.PaymentRepository

	clientsSrvOnce sync.Once
	clientsService service.ClientService
}

func newServiceProvider(cfg *config.Config, db *sqlx.DB, log *logger.Logger) *serviceProvider {
	return &serviceProvider{cfg: cfg, db: db, log: log}
}

func (sp *serviceProvider) TelegramUserRepo() repository.TelegramUserRepository {
	sp.telegramUserRepoOnce.Do(func() {
		sp.telegramUserRepo = telegramuser.New(sp.db)
	})
	return sp.telegramUserRepo
}

func (sp *serviceProvider) ClientRepo() repository.ClientRepository {
	sp.clientRepoOnce.Do(func() {
		sp.clientRepo = client.New(sp.db)
	})
	return sp.clientRepo
}

func (sp *serviceProvider) RentRepo() repository.RentRepository {
	sp.rentRepoOnce.Do(func() {
		sp.rentRepo = rent.New(sp.db)
	})
	return sp.rentRepo
}

func (sp *serviceProvider) ProductRepo() repository.ProductRepository {
	sp.productRepoOnce.Do(func() {
		sp.productRepo = product.New(sp.db)
	})
	return sp.productRepo
}

func (sp *serviceProvider) PaymentRepo() repository.PaymentRepository {
	sp.paymentRepoOnce.Do(func() {
		sp.paymentRepo = payment.New(sp.db)
	})
	return sp.paymentRepo
}

func (sp *serviceProvider) ClientsService() service.ClientService {
	if sp.clientsService == nil {
		sp.clientsService = client2.New(sp.db, sp.TelegramUserRepo(), sp.ClientRepo(), sp.log)
	}
	return sp.clientsService
}
