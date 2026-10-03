BINARY := app

.PHONY: test check build run
test:
	go test ./...

check:
	go vet ./...
	go test ./...

build:
	go build -o build/$(BINARY) ./cmd/bot

# Локальный запуск: переменные из deploy/env.example, BOT_TOKEN — тестового бота.
run: build
	./build/$(BINARY)
