package main

import (
	"context"
	"log/slog"
	"os"

	"apigofiberhorpug/config"
	"apigofiberhorpug/internal/bootstrap"
	"apigofiberhorpug/internal/database"
	deliveryhttp "apigofiberhorpug/internal/delivery/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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

	container := bootstrap.NewContainer(db, cfg.SecretKey)

	// HTTP Server
	app := fiber.New(fiber.Config{
		AppName:      "Horpug API v1.0",
		ErrorHandler: deliveryhttp.ErrorHandler,
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
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
	deliveryhttp.SetupRoutes(app, container.AuthUC, container.UserUC, container.RoleUC, container.PermUC, container.MenuUC, container.RoomUC, container.TenantUC, container.ContractUC, container.MeterReadingUC, container.BillUC, container.DashboardUC, container.AnalyticsUC, container.ExpenseUC, container.MaintenanceUC, container.PaymentUC, container.AnnouncementUC, container.ReportUC, container.ParkingUC, container.ParcelUC)

	slog.Info("เซิร์ฟเวอร์พร้อมทำงาน", "port", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		slog.Error("ไม่สามารถเปิดเซิร์ฟเวอร์ได้", "error", err)
		os.Exit(1)
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
