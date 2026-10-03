// Package config читает настройки из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Token string
	// DBPath в homelab задаёт compose: /data/<сервис>.db.
	DBPath   string
	Timezone *time.Location
	// GeminiAPIKey необязателен: без ключа сервис работает без модели.
	GeminiAPIKey string
	GeminiModel  string
	// GeminiBaseURL — корень API без /v1beta. В homelab это адрес шлюза учёта, иначе Google.
	GeminiBaseURL string
}

// Load собирает конфиг. Токен обязателен, остальное имеет значения по умолчанию.
func Load() (Config, error) {
	c := Config{
		Token:         os.Getenv("BOT_TOKEN"),
		DBPath:        envOr("BOT_DB", "app.db"),
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
		GeminiModel:   envOr("GEMINI_MODEL", "gemini-2.5-flash"),
		GeminiBaseURL: envOr("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com"),
	}
	if c.Token == "" {
		return c, errors.New("не задан BOT_TOKEN")
	}
	tz := envOr("BOT_TIMEZONE", "Asia/Bishkek")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return c, fmt.Errorf("неизвестная таймзона %q: %w", tz, err)
	}
	c.Timezone = loc
	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
