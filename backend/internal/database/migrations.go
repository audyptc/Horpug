package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// ApplyMigrations จะรัน migration SQL ที่ embed มากับ binary แบบเรียงลำดับ
func (db *DB) ApplyMigrations(ctx context.Context) error {
	if _, err := db.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`); err != nil {
		return fmt.Errorf("ไม่สามารถสร้างตาราง schema_migrations ได้: %v", err)
	}

	migrationPaths, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("ไม่สามารถอ่านรายการ migrations ได้: %v", err)
	}

	sort.Strings(migrationPaths)

	for _, migrationPath := range migrationPaths {
		version := strings.TrimSuffix(path.Base(migrationPath), path.Ext(migrationPath))

		var appliedVersion string
		err := db.Pool.QueryRow(ctx, `SELECT version FROM schema_migrations WHERE version = $1`, version).Scan(&appliedVersion)
		if err == nil {
			log.Printf("ℹ️  migration %s ถูกใช้งานแล้ว", version)
			continue
		}
		if err != pgx.ErrNoRows {
			return fmt.Errorf("ไม่สามารถตรวจสอบ migration %s ได้: %v", version, err)
		}

		migrationSQL, err := migrationFiles.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("ไม่สามารถอ่าน migration %s ได้: %v", version, err)
		}

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("ไม่สามารถเริ่ม transaction สำหรับ migration %s ได้: %v", version, err)
		}

		if _, err := tx.Exec(ctx, string(migrationSQL)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("ไม่สามารถรัน migration %s ได้: %v", version, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("ไม่สามารถบันทึก migration %s ได้: %v", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("ไม่สามารถ commit migration %s ได้: %v", version, err)
		}

		log.Printf("✅ migration %s สำเร็จ", version)
	}

	return nil
}
