# go-service-template

Заготовка Telegram-бота на Go для [homelab](https://github.com/IKulimanov/homelab): настройки из окружения,
SQLite с миграциями, клиент Gemini через шлюз учёта, сборка образа в ghcr.io при push в `main`.

## Новый сервис

1. На GitHub: «Use this template» → репозиторий с именем сервиса, например `habit-bot`. Имя образа в ghcr
   берётся из имени репозитория, `build.yml` править не нужно.
2. Клонировать рядом с homelab, в `~/dev/personal/<сервис>`.
3. Дописать логику в `internal/tg`, таблицы — новыми шагами в `internal/storage/db.go`.
4. Переменные — в `deploy/env.example`. Над обязательной переменной в комментарии слово «Обязателен».
5. Push в `main`, дождаться зелёной сборки и образа в ghcr.
6. В homelab: `scripts/new-service.sh <сервис>` (с `--llm 2`, если нужен Gemini с лимитом $2 в месяц), push.
   Через 5 минут служебный бот пришлёт ссылку на «Сейф» в панели: туда вписать `BOT_TOKEN`.

Модуль называется `service`. Переименовывать не обязательно; если хочется:

```bash
go mod edit -module habitbot && grep -rl '"service/' --include=*.go . | xargs sed -i '' 's|"service/|"habitbot/|'
```

## Что внутри

| Путь | Что |
|---|---|
| `cmd/bot` | запуск, версия сборки в логе, остановка по SIGTERM |
| `internal/config` | переменные окружения, проверка обязательных |
| `internal/storage` | SQLite: `pragma user_version` и шаги схемы, дописывать только в конец |
| `internal/llm` | `Generate` к Gemini; адрес — `GEMINI_BASE_URL`, по умолчанию Google |
| `internal/tg` | `/start` и `/help` |
| `deploy/env.example` | образец секретов для homelab |
| `Dockerfile` | копия `templates/Dockerfile` из homelab |

## Команды

```bash
make check
```

```bash
BOT_TOKEN=<токен тестового бота> make run
```
