package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"apihorpug/config"
	invoicerepository "apihorpug/internal/features/invoice/repository/postgres"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"
	menurepository "apihorpug/internal/features/menu/repository/postgres"
	"apihorpug/internal/http"
	"apihorpug/internal/platform/database"
	"apihorpug/internal/platform/filestore"
	"apihorpug/internal/platform/lineapi"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// @title Horpug API
// @version 1.0
// @description API for the Horpug dormitory management system.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	if err := menurepository.SeedMenus(db); err != nil {
		log.Fatalf("failed to seed menus: %v", err)
	}

	if err := database.SeedPermissions(db); err != nil {
		log.Fatalf("failed to seed permissions: %v", err)
	}

	if err := database.SeedAdmin(db, cfg); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	files, err := filestore.NewLocal(cfg.UploadDir)
	if err != nil {
		log.Fatalf("failed to prepare upload directory: %v", err)
	}

	// Backend is only reachable through nginx on the docker-internal network
	// (see docker-compose.yml — backend has no exposed port), so it's safe to
	// trust X-Forwarded-For from any private-range peer and recover the real
	// client IP nginx already forwards (nginx/proxy_params.conf).
	app := fiber.New(fiber.Config{
		ErrorHandler: http.ErrorHandler,
		// Uploaded files arrive as multipart bodies; the default 4 MB limit
		// would reject them. Kept in step with nginx's client_max_body_size.
		BodyLimit:   10 * 1024 * 1024,
		ProxyHeader: fiber.HeaderXForwardedFor,
		TrustProxy:  true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Private: true,
		},
	})
	// Turn a panic in any handler into a 500 instead of taking the whole
	// process (and every in-flight request) down with it.
	app.Use(recover.New())
	http.RegisterRoutes(app, db, cfg.SecretKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.CookieSecure, cfg.LineChannelAccessToken, cfg.LineChannelID, files)
	http.RegisterDocsRoutes(app)

	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("\n\x1b[1;32m➜\x1b[0m  \x1b[1mScalar UI:\x1b[0m  \x1b[36mhttp://localhost:%s/docs/scalar\x1b[0m\n", cfg.AppPort)
		fmt.Printf("\x1b[1;32m➜\x1b[0m  \x1b[1mSwagger UI:\x1b[0m \x1b[36mhttp://localhost:%s/docs/swagger\x1b[0m\n\n", cfg.AppPort)
	}()

	// docker stop / compose restarts send SIGTERM: stop accepting new
	// connections and let in-flight requests finish before exiting.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Hourly is plenty for a date-based status, and catches the day change
	// soon after midnight even though the server may have started at any hour.
	// The same run sends the weekly LINE reminders for overdue invoices.
	go invoiceusecase.RunInvoiceJobs(ctx, invoicerepository.NewRepository(db),
		lineapi.New(cfg.LineChannelAccessToken, cfg.LineChannelID), time.Hour)

	log.Printf("server running on :%s", cfg.AppPort)
	if err := app.Listen(":"+cfg.AppPort, fiber.ListenConfig{
		GracefulContext: ctx,
		ShutdownTimeout: 10 * time.Second,
	}); err != nil {
		log.Printf("server stopped: %v", err)
	}

	db.Close()
	log.Println("server shut down")
}
