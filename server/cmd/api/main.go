package main

import (
	"context"
	"log"

	"apigofiberhorpug/config"
	"apigofiberhorpug/internal/bootstrap"
	"apigofiberhorpug/internal/database"
	deliveryhttp "apigofiberhorpug/internal/delivery/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// @title  API
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ โหลด config ไม่สำเร็จ: %v", err)
	}

	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ เชื่อมต่อฐานข้อมูลไม่สำเร็จ: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(context.Background()); err != nil {
		log.Fatalf("❌ ไม่สามารถรัน migrations ได้: %v", err)
	}

	container := bootstrap.NewContainer(db, cfg.SecretKey)

	// HTTP Server
	app := fiber.New(fiber.Config{
		AppName:      "Horpug API v1.0",
		ErrorHandler: deliveryhttp.ErrorHandler,
	})
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})
	app.Get("/swagger", func(c fiber.Ctx) error {
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

	setupScalarDocs(app)
	deliveryhttp.SetupRoutes(app, container.AuthUC, container.UserUC, container.RoleUC, container.PermUC, container.MenuUC)

	log.Printf("🚀 เซิร์ฟเวอร์พร้อมทำงานแล้วที่พอร์ต %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("❌ ไม่สามารถเปิดเซิร์ฟเวอร์ได้: %v", err)
	}
}

func setupScalarDocs(app *fiber.App) {
	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})
	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(`<!doctype html>
<html>
  <head>
    <title>Horpug API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/docs/swagger.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`)
	})
}
