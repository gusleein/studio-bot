package converter

import (
	"encoding/json"
	"github.com/yourstudio/studio-bot/internal/domain"
	"github.com/yourstudio/studio-bot/internal/repository/appsettings/model"
)

func ToDomain(from model.AppSettings, to *domain.AppSettings) {
	switch from.Key {
	case "app_name":
		to.AppName = from.Value
	case "opening_hours":
		_ = json.Unmarshal([]byte(from.Value), &to.OpeningHours)
	case "custom_prices":
		_ = json.Unmarshal([]byte(from.Value), &to.CustomPrices)
	case "bot_admins":
		_ = json.Unmarshal([]byte(from.Value), &to.BotAdmins)
	}
}

func ToModelOpeningHours(from domain.OpeningHours) (to model.AppSettings) {
	to.Key = "opening_hours"
	val, _ := json.Marshal(from)
	to.Value = string(val)
	return
}

func ToModelCustomPrice(from []domain.CustomPrice) (to model.AppSettings) {
	to.Key = "custom_prices"
	val, _ := json.Marshal(from)
	to.Value = string(val)
	return
}

func ToModelBotAdmin(from []domain.BotAdmin) (to model.AppSettings) {
	to.Key = "bot_admins"
	val, _ := json.Marshal(from)
	to.Value = string(val)
	return
}
