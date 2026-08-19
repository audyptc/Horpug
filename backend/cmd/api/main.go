package main

import (
	"fmt"
	"log"
	"time"

	"apihorpug/config"
	menurepository "apihorpug/internal/features/menu/repository/postgres"
	"apihorpug/internal/http"
	"apihorpug/internal/platform/database"

	"github.com/gofiber/fiber/v3"
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

	// Backend is only reachable through nginx on the docker-internal network
	// (see docker-compose.yml — backend has no exposed port), so it's safe to
	// trust X-Forwarded-For from any private-range peer and recover the real
	// client IP nginx already forwards (nginx/proxy_params.conf).
	app := fiber.New(fiber.Config{
		ErrorHandler: http.ErrorHandler,
		ProxyHeader:  fiber.HeaderXForwardedFor,
		TrustProxy:   true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Private: true,
		},
	})
	http.RegisterRoutes(app, db, cfg.SecretKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.CookieSecure, cfg.LineChannelAccessToken, cfg.LineChannelID)
	http.RegisterDocsRoutes(app)

	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("\n\x1b[1;32m➜\x1b[0m  \x1b[1mScalar UI:\x1b[0m  \x1b[36mhttp://localhost:%s/docs/scalar\x1b[0m\n", cfg.AppPort)
		fmt.Printf("\x1b[1;32m➜\x1b[0m  \x1b[1mSwagger UI:\x1b[0m \x1b[36mhttp://localhost:%s/docs/swagger\x1b[0m\n\n", cfg.AppPort)
	}()

	log.Printf("server running on :%s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
