package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"apigofiberhorpug/config"
	"apigofiberhorpug/internal/bootstrap"
	"apigofiberhorpug/internal/database"
	deliveryhttp "apigofiberhorpug/internal/delivery/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// @title  API
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("โหลด config ไม่สำเร็จ", "error", err)
		os.Exit(1)
	}

	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("เชื่อมต่อฐานข้อมูลไม่สำเร็จ", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.ApplyMigrations(context.Background()); err != nil {
		slog.Error("ไม่สามารถรัน migrations ได้", "error", err)
		os.Exit(1)
	}

	container := bootstrap.NewContainer(db, cfg.SecretKey, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

	// HTTP Server
	app := fiber.New(fiber.Config{
		AppName:      "Horpug API v1.0",
		ErrorHandler: deliveryhttp.ErrorHandler,
		BodyLimit:    20 * 1024 * 1024, // 20MB for file uploads
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		if err := db.Pool.Ping(context.Background()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status": "down", "database": "disconnected",
			})
		}
		return c.JSON(fiber.Map{"status": "up", "database": "connected"})
	})

	app.Use("/uploads", static.New(cfg.UploadDir))

	setupScalarDocs(app)
	deliveryhttp.SetupRoutes(app, container, cfg)

	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("\n\x1b[1;32m➜\x1b[0m  \x1b[1mScalar UI:\x1b[0m  \x1b[36mhttp://localhost:%s/docs\x1b[0m\n", cfg.AppPort)
		fmt.Printf("\x1b[1;32m➜\x1b[0m  \x1b[1mSwagger UI:\x1b[0m \x1b[36mhttp://localhost:%s/swagger\x1b[0m\n\n", cfg.AppPort)
	}()

	slog.Info("เซิร์ฟเวอร์พร้อมทำงาน", "port", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		slog.Error("ไม่สามารถเปิดเซิร์ฟเวอร์ได้", "error", err)
		os.Exit(1)
	}
}
