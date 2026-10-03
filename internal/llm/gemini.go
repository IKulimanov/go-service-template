// Package llm — тонкий клиент Gemini API на net/http: SDK тянет много зависимостей ради одного вызова.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

type Gemini struct {
	key     string
	model   string
	baseURL string
	http    *http.Client
}

// NewGemini — клиент к корню API baseURL: Google или шлюз учёта homelab. Путь /v1beta добавляется здесь.
func NewGemini(key, model, baseURL string) *Gemini {
	return &Gemini{
		key:     key,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/") + "/v1beta",
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

type part struct {
	Text string `json:"text"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type request struct {
	Contents          []content `json:"contents"`
	SystemInstruction *content  `json:"systemInstruction,omitempty"`
}

type response struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
}

// Generate отправляет один запрос generateContent и возвращает текст ответа.
// Ключ идёт только заголовком: URL попадает в тексты ошибок net/http, а значит и в логи.
func (g *Gemini) Generate(ctx context.Context, system, user string) (string, error) {
	body := request{Contents: []content{{Role: "user", Parts: []part{{Text: user}}}}}
	if system != "" {
		body.SystemInstruction = &content{Parts: []part{{Text: system}}}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("gemini: собрать запрос: %w", err)
	}

	url := g.baseURL + "/models/" + neturl.PathEscape(g.model) + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("gemini: создать запрос: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.key)

	resp, err := g.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: отправить запрос: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("gemini: прочитать ответ, статус %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini: статус %d%s", resp.StatusCode, apiStatus(data))
	}

	var r response
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("gemini: разобрать ответ: %w", err)
	}
	if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: пустой ответ")
	}
	return r.Candidates[0].Content.Parts[0].Text, nil
}

// apiStatus достаёт из тела ошибки только код вида RESOURCE_EXHAUSTED: сообщение целиком
// в лог не несём — в нём может оказаться текст запроса.
func apiStatus(body []byte) string {
	var e struct {
		Error struct {
			Status string `json:"status"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) != nil || e.Error.Status == "" {
		return ""
	}
	return " " + e.Error.Status
}
