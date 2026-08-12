package http

import (
	"time"

	activityloghttp "apihorpug/internal/features/activitylog/delivery/http"
	activitylogrepository "apihorpug/internal/features/activitylog/repository/postgres"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	authhttp "apihorpug/internal/features/auth/delivery/http"
	authrepository "apihorpug/internal/features/auth/repository/postgres"
	authusecase "apihorpug/internal/features/auth/usecase"
	dormitoryhttp "apihorpug/internal/features/dormitory/delivery/http"
	dormitoryrepository "apihorpug/internal/features/dormitory/repository/postgres"
	dormitoryusecase "apihorpug/internal/features/dormitory/usecase"
	menuhttp "apihorpug/internal/features/menu/delivery/http"
	menurepository "apihorpug/internal/features/menu/repository/postgres"
	menuusecase "apihorpug/internal/features/menu/usecase"
	permissionhttp "apihorpug/internal/features/permission/delivery/http"
	permissiondomain "apihorpug/internal/features/permission/domain"
	permissionrepository "apihorpug/internal/features/permission/repository/postgres"
	permissionusecase "apihorpug/internal/features/permission/usecase"
	rolehttp "apihorpug/internal/features/role/delivery/http"
	rolerepository "apihorpug/internal/features/role/repository/postgres"
	roleusecase "apihorpug/internal/features/role/usecase"
	userhttp "apihorpug/internal/features/user/delivery/http"
	userrepository "apihorpug/internal/features/user/repository/postgres"
	userusecase "apihorpug/internal/features/user/usecase"
	"apihorpug/internal/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool, secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, cookieSecure bool) {
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
	dormitoryRepo := dormitoryrepository.NewRepository(db)
	dormitoryService := dormitoryusecase.New(dormitoryRepo)
	dormitoryHandler := dormitoryhttp.NewHandler(dormitoryService)
	activityLogRepo := activitylogrepository.NewRepository(db)
	activityLogService := activitylogusecase.New(activityLogRepo)
	activityLogHandler := activityloghttp.NewHandler(activityLogService)
	authTokenRepo := authrepository.NewRepository(db)
	authService := authusecase.New(userRepo, authTokenRepo, secretKey, accessTokenTTL, refreshTokenTTL)
	authHandler := authhttp.NewHandler(authService, cookieSecure)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	api := app.Group("/api/v1")
	api.Use(middleware.RequireAuth(secretKey))

	requirePermission := func(menuPath string, action permissiondomain.Action) fiber.Handler {
		return middleware.RequirePermission(db, menuPath, action)
	}

	api.Get("/permissions", requirePermission("/permissions", permissiondomain.ActionRead), permissionHandler.List)
	api.Post("/permissions", requirePermission("/permissions", permissiondomain.ActionCreate), permissionHandler.Create)

	api.Get("/menus", menuHandler.List)

	api.Get("/roles", requirePermission("/roles", permissiondomain.ActionRead), roleHandler.List)
	api.Get("/roles/:id", requirePermission("/roles", permissiondomain.ActionRead), roleHandler.Get)
	api.Post("/roles", requirePermission("/roles", permissiondomain.ActionCreate), roleHandler.Create)
	api.Put("/roles/:id", requirePermission("/roles", permissiondomain.ActionUpdate), roleHandler.Update)
	api.Delete("/roles/:id", requirePermission("/roles", permissiondomain.ActionDelete), roleHandler.Delete)

	api.Get("/users", requirePermission("/users", permissiondomain.ActionRead), userHandler.List)
	api.Get("/users/:id", requirePermission("/users", permissiondomain.ActionRead), userHandler.Get)
	api.Get("/users/:id/permissions", requirePermission("/users", permissiondomain.ActionRead), userHandler.GetPermissions)
	api.Post("/users", requirePermission("/users", permissiondomain.ActionCreate), userHandler.Create)
	api.Put("/users/:id", requirePermission("/users", permissiondomain.ActionUpdate), userHandler.Update)
	api.Delete("/users/:id", requirePermission("/users", permissiondomain.ActionDelete), userHandler.Delete)

	api.Get("/dormitories", requirePermission("/dormitories", permissiondomain.ActionRead), dormitoryHandler.List)
	api.Get("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionRead), dormitoryHandler.Get)
	api.Post("/dormitories", requirePermission("/dormitories", permissiondomain.ActionCreate), dormitoryHandler.Create)
	api.Put("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionUpdate), dormitoryHandler.Update)
	api.Delete("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionDelete), dormitoryHandler.Delete)

	api.Get("/activity-logs", requirePermission("/activity-logs", permissiondomain.ActionRead), activityLogHandler.List)
	api.Get("/activity-logs/:id", requirePermission("/activity-logs", permissiondomain.ActionRead), activityLogHandler.Get)
	api.Post("/activity-logs", requirePermission("/activity-logs", permissiondomain.ActionCreate), activityLogHandler.Create)
}
