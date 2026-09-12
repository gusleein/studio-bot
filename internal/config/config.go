package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

type BotConfig struct {
	Token   string
	Debug   bool
	Timeout int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type TributeConfig struct {
	APIKey     string
	BaseURL    string
	ShopID     uint64
	WebhookURL string
}

type StorageConfig struct {
	BasePath string
}

type Config struct {
	Env     string
	Server  ServerConfig
	DB      DBConfig
	Bot     BotConfig
	JWT     JWTConfig
	Tribute TributeConfig
	Storage StorageConfig
	Rent    RentConfig
	Admin   AdminConfig
}

// RentConfig — настройки клиентской аренды студии.
type RentConfig struct {
	AdminChatID  int64
	CardNumber   string
	PricePerHour int
	Timezone     string
}

type AdminConfig struct {
	Phone    string
	Username string
}

func Load(filenames ...string) Config {
	envFileName := cast.ToString(getOrReturnDefault("ENV_FILE_PATH", ".env"))
	if len(filenames) > 0 {
		envFileName = filenames[0]
	}

	if err := godotenv.Load(envFileName); err != nil {
		fmt.Println("No .env file found")
	}

	return Config{
		Env: cast.ToString(getOrReturnDefault("APP_ENV", EnvironmentDevelopment)),
		Server: ServerConfig{
			Host:         cast.ToString(getOrReturnDefault("SERVER_HOST", "0.0.0.0")),
			Port:         cast.ToInt(getOrReturnDefault("SERVER_PORT", 8088)),
			ReadTimeout:  cast.ToDuration(getOrReturnDefault("SERVER_READ_TIMEOUT", "10s")),
			WriteTimeout: cast.ToDuration(getOrReturnDefault("SERVER_WRITE_TIMEOUT", "10s")),
			IdleTimeout:  cast.ToDuration(getOrReturnDefault("SERVER_IDLE_TIMEOUT", "60s")),
		},
		DB: DBConfig{
			Host:            cast.ToString(getOrReturnDefault("DB_HOST", "localhost")),
			Port:            cast.ToInt(getOrReturnDefault("DB_PORT", 5432)),
			User:            cast.ToString(getOrReturnDefault("DB_USER", "studio")),
			Password:        cast.ToString(getOrReturnDefault("DB_PASSWORD", "studio")),
			DBName:          cast.ToString(getOrReturnDefault("DB_NAME", "studio_db")),
			SSLMode:         cast.ToString(getOrReturnDefault("DB_SSLMODE", "disable")),
			MaxOpenConns:    cast.ToInt(getOrReturnDefault("DB_MAX_OPEN_CONNS", 25)),
			MaxIdleConns:    cast.ToInt(getOrReturnDefault("DB_MAX_IDLE_CONNS", 10)),
			ConnMaxLifetime: cast.ToDuration(getOrReturnDefault("DB_CONN_MAX_LIFETIME", "5m")),
		},
		Bot: BotConfig{
			Token:   cast.ToString(getOrReturnDefault("TELEGRAM_BOT_TOKEN", "")),
			Debug:   cast.ToBool(getOrReturnDefault("TELEGRAM_BOT_DEBUG", false)),
			Timeout: cast.ToInt(getOrReturnDefault("TELEGRAM_BOT_TIMEOUT", 60)),
		},
		JWT: JWTConfig{
			Secret:     cast.ToString(getOrReturnDefault("JWT_SECRET", "")),
			Expiration: cast.ToDuration(getOrReturnDefault("JWT_EXPIRATION", "720h")),
		},
		Tribute: TributeConfig{
			APIKey:     tributeAPIKey(),
			BaseURL:    cast.ToString(getOrReturnDefault("TRIBUTE_BASE_URL", "https://tribute.tg/api/v1")),
			ShopID:     uint64(cast.ToInt(getOrReturnDefault("TRIBUTE_SHOP_ID", 0))),
			WebhookURL: cast.ToString(getOrReturnDefault("TRIBUTE_WEBHOOK_URL", "")),
		},
		Storage: StorageConfig{
			BasePath: cast.ToString(getOrReturnDefault("STORAGE_BASE_PATH", "./storage")),
		},
		Rent: RentConfig{
			AdminChatID:  int64(cast.ToInt(getOrReturnDefault("RENT_ADMIN_CHAT_ID", 576077782))),
			CardNumber:   cast.ToString(getOrReturnDefault("RENT_CARD_NUMBER", "")),
			PricePerHour: cast.ToInt(getOrReturnDefault("RENT_PRICE_PER_HOUR", 700)),
			Timezone:     cast.ToString(getOrReturnDefault("RENT_TIMEZONE", "Europe/Moscow")),
		},
		Admin: AdminConfig{
			Phone:    cast.ToString(getOrReturnDefault("ADMIN_PHONE", "")),
			Username: cast.ToString(getOrReturnDefault("ADMIN_USERNAME", "")),
		},
	}
}

func tributeAPIKey() string {
	if key := cast.ToString(getOrReturnDefault("TRIBUTE_API_KEY", "")); key != "" {
		return key
	}
	return cast.ToString(getOrReturnDefault("TRIBUTE_SECRET_KEY", ""))
}

func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	if _, exists := os.LookupEnv(key); exists {
		return os.Getenv(key)
	}
	return defaultValue
}
