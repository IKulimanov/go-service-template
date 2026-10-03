package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Клиент получает корень API, как у шлюза учёта homelab, сам добавляет /v1beta и шлёт ключ заголовком.
func TestGenerateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-2.5-flash:generateContent" {
			t.Errorf("путь %q", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("ключ не должен попадать в URL: %q", r.URL.RawQuery)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "secret" {
			t.Errorf("x-goog-api-key = %q", got)
		}
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Хорошие новости"}]}}]}`))
	}))
	defer srv.Close()

	got, err := NewGemini("secret", "gemini-2.5-flash", srv.URL+"/").Generate(context.Background(), "", "привет")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "Хорошие новости" {
		t.Errorf("ответ %q", got)
	}
}

// В ошибке — только статус API, без текста сообщения: там может оказаться запрос пользователя.
func TestGenerateErrorStatusOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"status":"RESOURCE_EXHAUSTED","message":"личный текст"}}`))
	}))
	defer srv.Close()

	_, err := NewGemini("k", "m", srv.URL).Generate(context.Background(), "", "x")
	if err == nil || !strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") || strings.Contains(err.Error(), "личный") {
		t.Fatalf("ошибка %v", err)
	}
}
