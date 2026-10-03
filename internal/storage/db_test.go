package storage

import (
	"context"
	"path/filepath"
	"testing"
)

// Повторное открытие не применяет шаги второй раз: иначе create/alter упали бы на уже существующей схеме.
func TestOpenMigratesOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	for range 2 {
		db, err := Open(path)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		var version int
		if err := db.QueryRow("pragma user_version").Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != len(steps()) {
			t.Errorf("user_version = %d, шагов %d", version, len(steps()))
		}
		db.Close()
	}
}

func TestRememberChatTwice(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	for range 2 {
		if err := db.RememberChat(ctx, 42); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := db.QueryRow("select count(*) from chat").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("чатов %d, ожидался 1", n)
	}
}
