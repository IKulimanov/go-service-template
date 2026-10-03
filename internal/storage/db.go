// Package storage — SQLite поверх database/sql, без внешнего инструмента миграций.
package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// steps — шаги схемы. Дописывать только в конец: номер применённого шага лежит
// в pragma user_version, поэтому переставлять и удалять шаги нельзя.
func steps() []string {
	return []string{schemaSQL}
}

type DB struct{ *sql.DB }

// Open открывает базу и доводит схему до актуальной. Файл создаётся, если его нет.
func Open(path string) (*DB, error) {
	dsn := "file:" + path +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(on)" +
		"&_pragma=synchronous(NORMAL)"

	sdb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("открыть базу %s: %w", path, err)
	}
	// SQLite всё равно пишет по одному; один коннект убирает SQLITE_BUSY внутри процесса.
	sdb.SetMaxOpenConns(1)
	if err := sdb.Ping(); err != nil {
		sdb.Close()
		return nil, fmt.Errorf("база недоступна: %w", err)
	}

	db := &DB{sdb}
	if err := db.migrate(); err != nil {
		sdb.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	var version int
	if err := db.QueryRow("pragma user_version").Scan(&version); err != nil {
		return fmt.Errorf("прочитать user_version: %w", err)
	}
	all := steps()
	for i := version; i < len(all); i++ {
		if err := db.applyStep(i, all[i]); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) applyStep(i int, body string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("шаг схемы %d: %w", i, err)
	}
	defer tx.Rollback()

	for _, stmt := range splitStatements(body) {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("шаг схемы %d, запрос %.60q: %w", i, stmt, err)
		}
	}
	// user_version не принимает плейсхолдеры, но i здесь — индекс среди наших же шагов.
	if _, err := tx.Exec(fmt.Sprintf("pragma user_version = %d", i+1)); err != nil {
		return fmt.Errorf("шаг схемы %d, отметка версии: %w", i, err)
	}
	return tx.Commit()
}

// splitStatements режет шаг по «;». Точки с запятой внутри строковых литералов в схеме не используются.
func splitStatements(body string) []string {
	var out []string
	for _, part := range strings.Split(body, ";") {
		if s := strings.TrimSpace(part); s != "" && !onlyComments(s) {
			out = append(out, s)
		}
	}
	return out
}

func onlyComments(s string) bool {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return false
		}
	}
	return true
}

// RememberChat отмечает чат, в котором нажали /start. Повторный /start ничего не меняет.
func (db *DB) RememberChat(ctx context.Context, chatID int64) error {
	_, err := db.ExecContext(ctx, `insert or ignore into chat (id) values (?)`, chatID)
	if err != nil {
		return fmt.Errorf("запомнить чат: %w", err)
	}
	return nil
}
