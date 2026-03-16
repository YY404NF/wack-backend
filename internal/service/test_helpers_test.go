package service

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"wack-backend/internal/database"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}
