package config

import "testing"

func TestLoadRequiresToken(t *testing.T) {
	t.Setenv("BOT_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatal("без BOT_TOKEN ожидалась ошибка")
	}
}

func TestLoadBadTimezone(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("BOT_TIMEZONE", "Mars/Olympus")
	if _, err := Load(); err == nil {
		t.Fatal("с неизвестной таймзоной ожидалась ошибка")
	}
}

// Адрес Gemini задаётся окружением: в homelab сервисы ходят в Gemini через шлюз учёта.
func TestGeminiBaseURL(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")

	t.Run("по умолчанию Google", func(t *testing.T) {
		t.Setenv("GEMINI_BASE_URL", "")
		c, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if c.GeminiBaseURL != "https://generativelanguage.googleapis.com" {
			t.Errorf("GeminiBaseURL = %q", c.GeminiBaseURL)
		}
	})

	t.Run("из окружения", func(t *testing.T) {
		t.Setenv("GEMINI_BASE_URL", "http://llm-gateway:8080")
		c, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if c.GeminiBaseURL != "http://llm-gateway:8080" {
			t.Errorf("GeminiBaseURL = %q", c.GeminiBaseURL)
		}
	})
}
