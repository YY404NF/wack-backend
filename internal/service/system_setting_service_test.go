package service

import (
	"path/filepath"
	"testing"
	"time"

	"wack-backend/internal/database"
	"wack-backend/internal/model"
)

func TestResolveActiveTermPrefersLatestStartedTerm(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wack.db")
	db, err := database.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	terms := []model.Term{
		{Name: "2025-2026-2", TermStartDate: "2026-03-02"},
		{Name: "2026-2027-1", TermStartDate: "2026-08-31"},
	}
	for _, term := range terms {
		if err := db.Create(&term).Error; err != nil {
			t.Fatalf("create term %s: %v", term.Name, err)
		}
	}

	service := NewSystemSettingService(db)
	now := time.Date(2026, time.March, 20, 13, 53, 0, 0, time.Local)

	term, err := service.resolveActiveTerm(now)
	if err != nil {
		t.Fatalf("resolve active term: %v", err)
	}
	if term.Name != "2025-2026-2" {
		t.Fatalf("expected active term 2025-2026-2, got %s", term.Name)
	}
}
