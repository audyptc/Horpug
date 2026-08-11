package http

import (
	menuhttp "apihorpug/internal/features/menu/delivery/http"
	menurepository "apihorpug/internal/features/menu/repository/postgres"
	menuusecase "apihorpug/internal/features/menu/usecase"
	permissionhttp "apihorpug/internal/features/permission/delivery/http"
	permissionrepository "apihorpug/internal/features/permission/repository/postgres"
	permissionusecase "apihorpug/internal/features/permission/usecase"
	rolehttp "apihorpug/internal/features/role/delivery/http"
	rolerepository "apihorpug/internal/features/role/repository/postgres"
	roleusecase "apihorpug/internal/features/role/usecase"
	userhttp "apihorpug/internal/features/user/delivery/http"
	userrepository "apihorpug/internal/features/user/repository/postgres"
	userusecase "apihorpug/internal/features/user/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool) {
	permissionRepo := permissionrepository.NewRepository(db)
	permissionService := permissionusecase.New(permissionRepo)
	permissionHandler := permissionhttp.NewHandler(permissionService)
	menuRepo := menurepository.NewRepository(db)
	menuService := menuusecase.New(menuRepo)
	menuHandler := menuhttp.NewHandler(menuService)
	roleRepo := rolerepository.NewRepository(db)
	roleService := roleusecase.New(roleRepo)
	roleHandler := rolehttp.NewHandler(roleService)
	userRepo := userrepository.NewRepository(db)
	userService := userusecase.New(userRepo)
	userHandler := userhttp.NewHandler(userService)

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
