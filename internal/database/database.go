package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var latestSchemaSQL = []string{
	`CREATE TABLE IF NOT EXISTS user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login_id TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		real_name TEXT NOT NULL,
		role INTEGER NOT NULL,
		managed_class_id INTEGER,
		status INTEGER NOT NULL DEFAULT 1,
		last_login_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_login_id ON user(login_id);`,
	`CREATE INDEX IF NOT EXISTS idx_role_status ON user(role, status);`,
	`CREATE INDEX IF NOT EXISTS idx_managed_class_id ON user(managed_class_id);`,
	`CREATE TABLE IF NOT EXISTS term (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		term_start_date TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE TABLE IF NOT EXISTS user_free_time (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		weekday INTEGER NOT NULL,
		section INTEGER NOT NULL,
		free_weeks TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT uk_term_user_time UNIQUE (term_id, user_id, weekday, section),
		CONSTRAINT fk_user_free_time_term FOREIGN KEY (term_id) REFERENCES term(id),
		CONSTRAINT fk_user_free_time_user FOREIGN KEY (user_id) REFERENCES user(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_user_free_time_term_user ON user_free_time(term_id, user_id);`,
	`CREATE TABLE IF NOT EXISTS class (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		grade INTEGER NOT NULL,
		major_name TEXT NOT NULL,
		class_name TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE INDEX IF NOT EXISTS idx_class_grade_major ON class(grade, major_name);`,
	`CREATE INDEX IF NOT EXISTS idx_class_grade_major_status ON class(grade, major_name, status);`,
	`CREATE INDEX IF NOT EXISTS idx_class_status ON class(status);`,
	`CREATE TABLE IF NOT EXISTS student (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_no TEXT NOT NULL UNIQUE,
		student_name TEXT NOT NULL,
		class_id INTEGER DEFAULT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
	`CREATE INDEX IF NOT EXISTS idx_student_class_id ON student(class_id);`,
	`CREATE INDEX IF NOT EXISTS idx_student_class_status ON student(class_id, status);`,
	`CREATE INDEX IF NOT EXISTS idx_student_status ON student(status);`,
	`CREATE INDEX IF NOT EXISTS idx_student_name ON student(student_name);`,
	`CREATE TABLE IF NOT EXISTS course (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		grade INTEGER NOT NULL,
		course_name TEXT NOT NULL,
		teacher_name TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_course_term FOREIGN KEY (term_id) REFERENCES term(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_course_term_grade ON course(term_id, grade);`,
	`CREATE INDEX IF NOT EXISTS idx_course_term_course_name ON course(term_id, course_name);`,
	`CREATE INDEX IF NOT EXISTS idx_course_term_teacher_name ON course(term_id, teacher_name);`,
	`CREATE INDEX IF NOT EXISTS idx_course_term_status ON course(term_id, status);`,
	`CREATE TABLE IF NOT EXISTS course_group (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		course_id INTEGER NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_course_group_term FOREIGN KEY (term_id) REFERENCES term(id),
		CONSTRAINT fk_course_group_course FOREIGN KEY (course_id) REFERENCES course(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_term_course_id ON course_group(term_id, course_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_course_id ON course_group(course_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_term_course_status ON course_group(term_id, course_id, status);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_status ON course_group(status);`,
	`CREATE TABLE IF NOT EXISTS course_group_student (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		course_group_id INTEGER NOT NULL,
		student_id INTEGER NOT NULL,
		class_id INTEGER DEFAULT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT uk_term_group_student UNIQUE (term_id, course_group_id, student_id),
		CONSTRAINT fk_course_group_student_term FOREIGN KEY (term_id) REFERENCES term(id),
		CONSTRAINT fk_course_group_student_group FOREIGN KEY (course_group_id) REFERENCES course_group(id),
		CONSTRAINT fk_course_group_student_student FOREIGN KEY (student_id) REFERENCES student(id),
		CONSTRAINT fk_course_group_student_class FOREIGN KEY (class_id) REFERENCES class(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_group_id ON course_group_student(term_id, course_group_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_student_id ON course_group_student(term_id, student_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_class_id ON course_group_student(term_id, class_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_group_class_id ON course_group_student(term_id, course_group_id, class_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_group_status ON course_group_student(term_id, course_group_id, status);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_student_term_student_status ON course_group_student(term_id, student_id, status);`,
	`CREATE TABLE IF NOT EXISTS course_group_lesson (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		course_group_id INTEGER NOT NULL,
		week_no INTEGER NOT NULL,
		weekday INTEGER NOT NULL,
		section INTEGER NOT NULL,
		building_name TEXT NOT NULL,
		room_name TEXT NOT NULL,
		status INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT uk_term_group_lesson_time UNIQUE (term_id, course_group_id, week_no, weekday, section),
		CONSTRAINT fk_course_group_lesson_term FOREIGN KEY (term_id) REFERENCES term(id),
		CONSTRAINT fk_course_group_lesson_group FOREIGN KEY (course_group_id) REFERENCES course_group(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_lesson_time_room ON course_group_lesson(term_id, week_no, weekday, section, building_name, room_name);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_lesson_term_group_id ON course_group_lesson(term_id, course_group_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_lesson_course_group_id ON course_group_lesson(course_group_id);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_lesson_term_group_status ON course_group_lesson(term_id, course_group_id, status);`,
	`CREATE INDEX IF NOT EXISTS idx_course_group_lesson_status ON course_group_lesson(status);`,
	`CREATE TABLE IF NOT EXISTS attendance_record (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		course_id INTEGER NOT NULL,
		course_group_lesson_id INTEGER NOT NULL,
		student_id INTEGER NOT NULL,
		class_id INTEGER,
		attendance_status INTEGER NOT NULL,
		updated_by_user_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT uk_lesson_student UNIQUE (course_group_lesson_id, student_id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_student ON attendance_record(term_id, student_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_class ON attendance_record(term_id, class_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_course ON attendance_record(term_id, course_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_student_status ON attendance_record(term_id, student_id, attendance_status);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_class_status ON attendance_record(term_id, class_id, attendance_status);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_course_status ON attendance_record(term_id, course_id, attendance_status);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_lesson_id ON attendance_record(term_id, course_group_lesson_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_term_updated_by_user_id ON attendance_record(term_id, updated_by_user_id);`,
	`CREATE TABLE IF NOT EXISTS attendance_record_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		term_id INTEGER NOT NULL,
		attendance_record_id INTEGER NOT NULL,
		operated_by_user_id INTEGER NOT NULL,
		old_attendance_status INTEGER DEFAULT NULL,
		new_attendance_status INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_attendance_record_log_term FOREIGN KEY (term_id) REFERENCES term(id),
		CONSTRAINT fk_attendance_record_log_record FOREIGN KEY (attendance_record_id) REFERENCES attendance_record(id),
		CONSTRAINT fk_attendance_record_log_user FOREIGN KEY (operated_by_user_id) REFERENCES user(id)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_id ON attendance_record_log(term_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_attendance_record_id ON attendance_record_log(term_id, attendance_record_id);`,
	`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_operated_by_user_id ON attendance_record_log(term_id, operated_by_user_id);`,
}

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

	if err := ensureLatestSchema(db); err != nil {
		return nil, fmt.Errorf("ensure latest schema: %w", err)
	}

	return db, nil
}

func ensureLatestSchema(db *gorm.DB) error {
	for _, stmt := range latestSchemaSQL {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	if err := ensureAttendanceRecordLogOldStatusNullable(db); err != nil {
		return err
	}
	if err := normalizeAttendanceRecordLogCreateRows(db); err != nil {
		return err
	}
	return nil
}

type sqliteTableColumn struct {
	Name    string `gorm:"column:name"`
	NotNull int    `gorm:"column:notnull"`
}

func ensureAttendanceRecordLogOldStatusNullable(db *gorm.DB) error {
	var columns []sqliteTableColumn
	if err := db.Raw("PRAGMA table_info(attendance_record_log)").Scan(&columns).Error; err != nil {
		return err
	}
	for _, column := range columns {
		if column.Name == "old_attendance_status" && column.NotNull != 0 {
			return rebuildAttendanceRecordLogTable(db)
		}
	}
	return nil
}

func rebuildAttendanceRecordLogTable(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}
	defer conn.ExecContext(ctx, "PRAGMA foreign_keys = ON")

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	stmts := []string{
		`CREATE TABLE attendance_record_log__new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			term_id INTEGER NOT NULL,
			attendance_record_id INTEGER NOT NULL,
			operated_by_user_id INTEGER NOT NULL,
			old_attendance_status INTEGER DEFAULT NULL,
			new_attendance_status INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_attendance_record_log_term FOREIGN KEY (term_id) REFERENCES term(id),
			CONSTRAINT fk_attendance_record_log_record FOREIGN KEY (attendance_record_id) REFERENCES attendance_record(id),
			CONSTRAINT fk_attendance_record_log_user FOREIGN KEY (operated_by_user_id) REFERENCES user(id)
		);`,
		`INSERT INTO attendance_record_log__new (
			id,
			term_id,
			attendance_record_id,
			operated_by_user_id,
			old_attendance_status,
			new_attendance_status,
			created_at
		)
		SELECT
			id,
			term_id,
			attendance_record_id,
			operated_by_user_id,
			CASE
				WHEN old_attendance_status = new_attendance_status THEN NULL
				ELSE old_attendance_status
			END AS old_attendance_status,
			new_attendance_status,
			created_at
		FROM attendance_record_log;`,
		`DROP TABLE attendance_record_log;`,
		`ALTER TABLE attendance_record_log__new RENAME TO attendance_record_log;`,
		`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_id ON attendance_record_log(term_id);`,
		`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_attendance_record_id ON attendance_record_log(term_id, attendance_record_id);`,
		`CREATE INDEX IF NOT EXISTS idx_attendance_record_log_term_operated_by_user_id ON attendance_record_log(term_id, operated_by_user_id);`,
	}
	for _, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func normalizeAttendanceRecordLogCreateRows(db *gorm.DB) error {
	return db.Exec(`
		UPDATE attendance_record_log
		SET old_attendance_status = NULL
		WHERE old_attendance_status IS NOT NULL
		  AND old_attendance_status = new_attendance_status
	`).Error
}
