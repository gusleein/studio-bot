package appsettings

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository"
	"github.com/yourstudio/studio-bot/internal/repository/appsettings/converter"
	"github.com/yourstudio/studio-bot/internal/repository/appsettings/model"
	"github.com/yourstudio/studio-bot/internal/service"
	"github.com/yourstudio/studio-bot/pkg/logger"
)

type Service struct {
	db              *sqlx.DB
	appSettingsRepo repository.AppSettingsRepository
	log             *logger.Logger
}

var _ service.AppSettingsService = (*Service)(nil)

func New(
	db *sqlx.DB,
	appSettingsRepo repository.AppSettingsRepository,
	log *logger.Logger,
) *Service {
	return &Service{
		db:              db,
		appSettingsRepo: appSettingsRepo,
		log:             log,
	}
}

func (s *Service) Get(ctx context.Context) (result domain.AppSettings, err error) {
	list, err := s.appSettingsRepo.GetAllAppSettings(ctx)
	if err != nil {
		return
	}
	for _, setting := range list {
		converter.ToDomain(setting, &result)
	}
	return
}

func (s *Service) UpdateAppName(ctx context.Context, appName string) (err error) {
	return s.appSettingsRepo.UpdateAppSettings(ctx, model.AppSettings{Key: "app_name", Value: appName})
}

func (s *Service) UpdateOpeningHours(ctx context.Context, openingHours domain.OpeningHours) (err error) {
	return s.appSettingsRepo.UpdateAppSettings(ctx, converter.ToModelOpeningHours(openingHours))
}

func (s *Service) UpdateCustomPrices(ctx context.Context, customPrices []domain.CustomPrice) (err error) {
	return s.appSettingsRepo.UpdateAppSettings(ctx, converter.ToModelCustomPrice(customPrices))
}

func (s *Service) UpdateBotAdmins(ctx context.Context, botAdmins []domain.BotAdmin) (err error) {
	return s.appSettingsRepo.UpdateAppSettings(ctx, converter.ToModelBotAdmin(botAdmins))
}
