package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"apigofiberhorpug/internal/database" // นำเข้าแพ็กเกจสำหรับเชื่อมต่อฐานข้อมูล

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/spf13/viper"
)

func main() {
	// 1. โหลดค่า config จาก .env และเปิดการอ่าน Environment Variables จาก OS ทับค่าได้
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("⚠️  ไม่พบไฟล์ .env ระบบจะใช้ Environment Variables จาก OS")
		} else {
			log.Printf("⚠️  โหลดไฟล์ config ไม่สำเร็จ: %v", err)
		}
	}

	// 2. ดึงค่าจาก Env ตามโครงสร้าง APP_/DB_
	host := viper.GetString("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	dbPort := viper.GetString("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	username := viper.GetString("DB_USERNAME")
	if username == "" {
		username = "postgres"
	}

	password := viper.GetString("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	dbName := viper.GetString("DB_NAME")
	if dbName == "" {
		dbName = "goapi_db"
	}

	sslMode := viper.GetString("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	// รองรับรหัสผ่านที่มีอักขระพิเศษ เช่น ! หรือ @
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		username,
		url.QueryEscape(password),
		host,
		dbPort,
		dbName,
		sslMode,
	)

	port := viper.GetString("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// 3. เชื่อมต่อฐานข้อมูล PostgreSQL
	db, err := database.ConnectPostgres(dbURL)
	if err != nil {
		log.Fatalf("❌ บูตระบบไม่สำเร็จเนื่องจากปัญหาฐานข้อมูล: %v", err)
	}
	defer db.Close() // ปิดสายเมื่อแอปพลิเคชันปิดตัวลง

	if err := db.ApplyMigrations(context.Background()); err != nil {
		log.Fatalf("❌ ไม่สามารถรัน migrations ได้: %v", err)
	}

	// 4. บูต Fiber App พร้อมตั้งค่าพื้นฐาน
	app := fiber.New(fiber.Config{
		AppName: "GoAPI v1.0",
	})

	// ใส่ Middlewares ยอดนิยม
	app.Use(logger.New())  // แสดง Log ของ Request ที่วิ่งเข้ามาบนหน้าจอ Console
	app.Use(recover.New()) // ป้องกันไม่ให้แอปพลิเคชันแครช (Panic) เวลาเกิด Error

	// 5. สร้าง Route พื้นฐานสำหรับตรวจสอบสถานะระบบ
	app.Get("/health", func(c fiber.Ctx) error {
		// ทดสอบส่ง Context ไปเช็คกับ DB สั้นๆ
		err := db.Pool.Ping(context.Background())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":   "down",
				"database": "disconnected",
			})
		}
		return c.JSON(fiber.Map{
			"status":   "up",
			"database": "connected",
		})
	})

	// 6. สร้าง Route สำหรับ Scalar API Reference Documentation
	setupScalarDocs(app)

	// 7. รันเซิร์ฟเวอร์
	log.Printf("🚀 เซิร์ฟเวอร์พร้อมทำงานแล้วที่พอร์ต %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ ไม่สามารถเปิดเซิร์ฟเวอร์ได้: %v", err)
	}
}

// setupScalarDocs วางทางผ่านสำหรับเข้าไปหน้าอ่านเอกสาร API สวยๆ
func setupScalarDocs(app *fiber.App) {
	// เสิร์ฟไฟล์ JSON ของ Swagger ให้ระบบรับทราบ
	// (เมื่อคุณรัน swag init ไฟล์นี้จะไปอยู่ที่ ./docs/swagger.json)
	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// ส่องหน้าเว็บเอกสารผ่านทาง Browser: http://localhost:8080/docs
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(`
			<!doctype html>
			<html>
			  <head>
			    <title>GoAPI Reference Docs</title>
			    <meta charset="utf-8" />
			    <meta name="viewport" content="width=device-width, initial-scale=1" />
			  </head>
			  <body>
			    <!-- ระบุปลายทางไฟล์ JSON ที่เราวางไว้ด้านบน -->
			    <script id="api-reference" data-url="/docs/swagger.json"></script>
			    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
			  </body>
			</html>
		`)
	})
}
