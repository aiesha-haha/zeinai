package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
	Gemini   GeminiConfig
	FCM      FCMConfig
	WebPush  WebPushConfig
	Mobizon  MobizonConfig
	SMTP     SMTPConfig
	OpenAI   OpenAIConfig
	Kimi     KimiConfig
}

type AppConfig struct {
	Env      string
	Port     string
	LogLevel string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	UI_Port  string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type GeminiConfig struct {
	APIKey string
}

type OpenAIConfig struct {
	APIKey string
}

type KimiConfig struct {
	APIKey string
}

type FCMConfig struct {
	CredentialsPath string
}

type WebPushConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Subscriber      string
}

type MobizonConfig struct {
	APIKey string
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func LoadConfig() *Config {
	// Only load .env file if it exists (local dev), otherwise rely on env vars (docker)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file found but could not be loaded")
		}
	}

	return &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "local"),
			Port:     getEnv("APP_PORT", "8000"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "zeinai"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			UI_Port:  getEnv("RABBITMQ_UI_PORT", "15672"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "secret"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 72*time.Hour),
		},
		Gemini: GeminiConfig{
			APIKey: getEnv("GEMINI_API_KEY", ""),
		},
		FCM: FCMConfig{
			CredentialsPath: getEnv("FCM_CREDENTIALS_PATH", ""),
		},
		WebPush: WebPushConfig{
			VAPIDPublicKey:  getEnv("WEBPUSH_VAPID_PUBLIC_KEY", ""),
			VAPIDPrivateKey: getEnv("WEBPUSH_VAPID_PRIVATE_KEY", ""),
			Subscriber:      getEnv("WEBPUSH_SUBSCRIBER", "mailto:noreply@zeinai.kz"),
		},
		Mobizon: MobizonConfig{
			APIKey: getEnv("MOBIZON_API_KEY", ""),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "ZeinAI <noreply@zeinai.kz>"),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
		},
		Kimi: KimiConfig{
			APIKey: getEnv("KIMI_API_KEY", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return fallback
}
