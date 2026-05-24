package database

import (
	"context"
	"fmt"
	"log"
	"time"

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
		return nil, fmt.Errorf("เชื่อมต่อ Postgres สำเร็จแต่ Ping ไม่ผ่าน: %v", err)
	}

	log.Println("🐘 เชื่อมต่อ PostgreSQL สำเร็จเรียบร้อยแล้ว!")
	return &DB{Pool: pool}, nil
}

// Close ปิด Pool เมื่อแอปพลิเคชันจบการทำงาน
func (db *DB) Close() {
	db.Pool.Close()
}
