package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB พลิกแพลงให้เก็บ *pgxpool.Pool เอาไว้ใช้งานทั่วทั้งแอป
type DB struct {
	Pool *pgxpool.Pool
}

// ConnectPostgres ทำหน้าที่ลุยเชื่อมต่อฐานข้อมูลตามคู่สายที่เราส่งให้
func ConnectPostgres(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := newPostgresPool(ctx, databaseURL)
	if err != nil {
		if isDatabaseDoesNotExist(err) {
			if createErr := createDatabaseIfMissing(ctx, databaseURL); createErr != nil {
				return nil, createErr
			}

			pool, err = newPostgresPool(ctx, databaseURL)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	log.Println("🐘 เชื่อมต่อ PostgreSQL สำเร็จเรียบร้อยแล้ว!")
	return &DB{Pool: pool}, nil
}

func newPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// 1. ตั้งค่า Config สำหรับ Connection Pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถอ่านรูปแบบ Database URL ได้: %v", err)
	}

	// ปรับแต่งตามใจชอบ (เช่น เชื่อมต่อค้างไว้ขั้นต่ำเท่าไหร่, สูงสุดเท่าไหร่)
	config.MaxConns = 25
	config.MinConns = 5

	// 2. เริ่มทำการเปิด Pool เชื่อมต่อไปยัง Postgres
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถสร้าง Connection Pool ได้: %v", err)
	}

	// 3. ทดสอบการเชื่อมต่อจริง (Ping)
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("เชื่อมต่อ Postgres สำเร็จแต่ Ping ไม่ผ่าน: %v", err)
	}

	return pool, nil
}

func createDatabaseIfMissing(ctx context.Context, databaseURL string) error {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("ไม่สามารถแยกข้อมูล Database URL ได้: %v", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return fmt.Errorf("ไม่สามารถสร้างฐานข้อมูลได้: ไม่พบชื่อ database ใน URL")
	}

	maintenanceURL := *parsedURL
	maintenanceURL.Path = "/postgres"

	conn, err := pgx.Connect(ctx, maintenanceURL.String())
	if err != nil {
		return fmt.Errorf("ไม่สามารถเชื่อมต่อฐานข้อมูล maintenance เพื่อสร้าง database ได้: %v", err)
	}
	defer conn.Close(ctx)

	quotedDBName := strings.ReplaceAll(dbName, `"`, `""`)
	query := fmt.Sprintf(`CREATE DATABASE "%s"`, quotedDBName)
	if _, err := conn.Exec(ctx, query); err != nil {
		if isDatabaseDoesNotExist(err) {
			return fmt.Errorf("ไม่สามารถสร้างฐานข้อมูล %s ได้: %v", dbName, err)
		}

		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			log.Printf("ℹ️  database %s มีอยู่แล้ว", dbName)
			return nil
		}

		return fmt.Errorf("ไม่สามารถสร้างฐานข้อมูล %s ได้: %v", dbName, err)
	}

	log.Printf("✅ สร้างฐานข้อมูล %s เรียบร้อยแล้ว", dbName)
	return nil
}

func isDatabaseDoesNotExist(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "3D000"
	}

	return strings.Contains(strings.ToLower(err.Error()), "does not exist")
}

// Close ปิด Pool เมื่อแอปพลิเคชันจบการทำงาน
func (db *DB) Close() {
	db.Pool.Close()
}
