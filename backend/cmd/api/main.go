package main

import (
	"log"

	"apihorpug/config"
	"apihorpug/internal/http"
	"apihorpug/internal/platform/database"

	"github.com/gofiber/fiber/v3"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	if err := database.SeedMenus(db); err != nil {
		log.Fatalf("failed to seed menus: %v", err)
	}

	if err := database.SeedPermissions(db); err != nil {
		log.Fatalf("failed to seed permissions: %v", err)
	}

	app := fiber.New(fiber.Config{ErrorHandler: http.ErrorHandler})
	http.RegisterRoutes(app, db)

	log.Printf("server running on :%s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
