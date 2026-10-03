package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	// Встроенная база таймзон: в образе distroless её нет.
	_ "time/tzdata"

	"service/internal/config"
	"service/internal/storage"
	"service/internal/tg"
)

// version подставляет сборка: -ldflags "-X main.version=sha-…".
var version = "dev"

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("конфигурация", "err", err)
		os.Exit(1)
	}

	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Error("база данных", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	app, err := tg.New(cfg, db, log)
	if err != nil {
		log.Error("телеграм", "err", err)
		os.Exit(1)
	}

	// SIGTERM приходит от docker stop; compose homelab ждёт до 90 с, потом убивает процесс.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("запущен", "version", version, "db", cfg.DBPath, "gemini", cfg.GeminiAPIKey != "")
	app.Run(ctx)
	log.Info("остановлен")
}
