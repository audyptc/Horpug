package http

import (
	menuhttp "apihorpug/internal/features/menu/delivery/http"
	permissionhttp "apihorpug/internal/features/permission/delivery/http"
	rolehttp "apihorpug/internal/features/role/delivery/http"
	userhttp "apihorpug/internal/features/user/delivery/http"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool) {
	permissionHandler := permissionhttp.NewHandler(db)
	menuHandler := menuhttp.NewHandler(db)
	roleHandler := rolehttp.NewHandler(db)
	userHandler := userhttp.NewHandler(db)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	api.Get("/permissions", permissionHandler.List)
	api.Post("/permissions", permissionHandler.Create)

	api.Get("/menus", menuHandler.List)

	api.Get("/roles", roleHandler.List)
	api.Get("/roles/:id", roleHandler.Get)
	api.Post("/roles", roleHandler.Create)
	api.Put("/roles/:id", roleHandler.Update)
	api.Delete("/roles/:id", roleHandler.Delete)

	api.Get("/users", userHandler.List)
	api.Get("/users/:id", userHandler.Get)
	api.Get("/users/:id/permissions", userHandler.GetPermissions)
	api.Post("/users", userHandler.Create)
	api.Put("/users/:id", userHandler.Update)
	api.Delete("/users/:id", userHandler.Delete)
}
