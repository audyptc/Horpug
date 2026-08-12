package main

import (
	"log"

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

	app := fiber.New(fiber.Config{ErrorHandler: http.ErrorHandler})
	http.RegisterRoutes(app, db, cfg.SecretKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.CookieSecure)
	http.RegisterDocsRoutes(app)

	log.Printf("server running on :%s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
