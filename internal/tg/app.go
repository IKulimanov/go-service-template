// Package tg — слой Telegram: команды и ответы.
package tg

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"service/internal/config"
	"service/internal/llm"
	"service/internal/storage"
)

const helpText = `Это заготовка сервиса homelab.
/start — начать
/help — эта подсказка`

type App struct {
	db  *storage.DB
	b   *bot.Bot
	log *slog.Logger
	// llm — nil, если ключа Gemini нет.
	llm *llm.Gemini
}

func New(cfg config.Config, db *storage.DB, log *slog.Logger) (*App, error) {
	a := &App{db: db, log: log}
	if cfg.GeminiAPIKey != "" {
		a.llm = llm.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiBaseURL)
	}
	b, err := bot.New(cfg.Token,
		bot.WithDefaultHandler(func(context.Context, *bot.Bot, *models.Update) {}),
		bot.WithAllowedUpdates(bot.AllowedUpdates{"message"}),
	)
	if err != nil {
		return nil, fmt.Errorf("создать бота: %w", err)
	}
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, a.start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommandStartOnly, a.help)
	a.b = b
	return a, nil
}

// Run слушает Telegram, пока не отменён ctx.
func (a *App) Run(ctx context.Context) { a.b.Start(ctx) }

func (a *App) start(ctx context.Context, _ *bot.Bot, u *models.Update) {
	if err := a.db.RememberChat(ctx, u.Message.Chat.ID); err != nil {
		a.log.Error("старт", "err", err)
	}
	a.reply(ctx, u.Message.Chat.ID, "Привет! "+helpText)
}

func (a *App) help(ctx context.Context, _ *bot.Bot, u *models.Update) {
	a.reply(ctx, u.Message.Chat.ID, helpText)
}

func (a *App) reply(ctx context.Context, chatID int64, text string) {
	if _, err := a.b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text}); err != nil {
		a.log.Error("ответ не отправлен", "err", err)
	}
}
