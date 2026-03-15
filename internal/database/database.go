package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wack-backend/internal/model"
)

func OpenAndMigrate(databasePath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir db dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Class{},
		&model.UserClass{},
		&model.StudentFreeTime{},
		&model.Course{},
		&model.CourseStudent{},
		&model.CourseClass{},
		&model.CourseSession{},
		&model.AttendanceCheck{},
		&model.AttendanceDetail{},
		&model.AttendanceDetailLog{},
		&model.AdminOperationLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}
