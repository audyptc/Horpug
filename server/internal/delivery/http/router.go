package http

import (
	"apigofiberhorpug/internal/delivery/http/middleware"
	v1 "apigofiberhorpug/internal/delivery/http/v1"
	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(
	app *fiber.App,
	auth *usecase.AuthUseCase,
	users *usecase.UserUseCase,
	roles *usecase.RoleUseCase,
	perms *usecase.PermissionUseCase,
	menus *usecase.MenuUseCase,
	rooms *usecase.RoomUseCase,
	tenants *usecase.TenantUseCase,
	contracts *usecase.ContractUseCase,
	meters *usecase.MeterReadingUseCase,
) {
	authH := v1.NewAuthHandler(auth)
	userH := v1.NewUserHandler(users)
	roleH := v1.NewRoleHandler(roles)
	permH := v1.NewPermissionHandler(perms)
	menuH := v1.NewMenuHandler(menus)
	roomH := v1.NewRoomHandler(rooms)
	tenantH := v1.NewTenantHandler(tenants)
	contractH := v1.NewContractHandler(contracts)
	meterH := v1.NewMeterReadingHandler(meters)

	api := app.Group("/api/v1")

	// Public: auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authH.Login)
	authGroup.Post("/refresh", authH.Refresh)
	authGroup.Post("/logout", authH.Logout)

	// Protected: all routes below require a valid JWT
	protected := api.Group("", middleware.RequireAuth(auth))

	// Users
	usersGroup := protected.Group("/users")
	usersGroup.Get("/", middleware.RequirePermission("users.read"), userH.List)
	usersGroup.Post("/", middleware.RequirePermission("users.create"), userH.Create)
	usersGroup.Get("/:id", middleware.RequirePermission("users.read"), userH.GetByID)
	usersGroup.Put("/:id", middleware.RequirePermission("users.update"), userH.Update)
	usersGroup.Delete("/:id", middleware.RequirePermission("users.delete"), userH.Delete)
	usersGroup.Put("/:id/role", middleware.RequirePermission("users.update"), userH.AssignRole)

	// Roles
	rolesGroup := protected.Group("/roles")
	rolesGroup.Get("/", middleware.RequirePermission("roles.read"), roleH.List)
	rolesGroup.Get("/active", middleware.RequirePermission("roles.read"), roleH.ListActive)
	rolesGroup.Post("/", middleware.RequirePermission("roles.create"), roleH.Create)
	rolesGroup.Get("/:id", middleware.RequirePermission("roles.read"), roleH.GetByID)
	rolesGroup.Put("/:id", middleware.RequirePermission("roles.update"), roleH.Update)
	rolesGroup.Delete("/:id", middleware.RequirePermission("roles.delete"), roleH.Delete)
	rolesGroup.Put("/:id/permissions", middleware.RequirePermission("roles.update"), roleH.AssignPermissions)

	// Permissions (read-only, managed via migrations)
	permGroup := protected.Group("/permissions")
	permGroup.Get("/", middleware.RequirePermission("permissions.read"), permH.List)

	// Menus
	menusGroup := protected.Group("/menus")
	menusGroup.Get("/", middleware.RequirePermission("menus.read"), menuH.List)
	menusGroup.Get("/:id", middleware.RequirePermission("menus.read"), menuH.GetByID)

	// Rooms
	roomsGroup := protected.Group("/rooms")
	roomsGroup.Get("/", middleware.RequirePermission("rooms.read"), roomH.List)
	roomsGroup.Post("/", middleware.RequirePermission("rooms.create"), roomH.Create)
	roomsGroup.Get("/:id", middleware.RequirePermission("rooms.read"), roomH.GetByID)
	roomsGroup.Put("/:id", middleware.RequirePermission("rooms.update"), roomH.Update)
	roomsGroup.Delete("/:id", middleware.RequirePermission("rooms.delete"), roomH.Delete)

	// Tenants
	tenantsGroup := protected.Group("/tenants")
	tenantsGroup.Get("/", middleware.RequirePermission("tenants.read"), tenantH.List)
	tenantsGroup.Post("/", middleware.RequirePermission("tenants.create"), tenantH.Create)
	tenantsGroup.Get("/:id", middleware.RequirePermission("tenants.read"), tenantH.GetByID)
	tenantsGroup.Put("/:id", middleware.RequirePermission("tenants.update"), tenantH.Update)
	tenantsGroup.Delete("/:id", middleware.RequirePermission("tenants.delete"), tenantH.Delete)

	// Contracts
	contractsGroup := protected.Group("/contracts")
	contractsGroup.Get("/", middleware.RequirePermission("contracts.read"), contractH.List)
	contractsGroup.Post("/", middleware.RequirePermission("contracts.create"), contractH.Create)
	contractsGroup.Get("/:id", middleware.RequirePermission("contracts.read"), contractH.GetByID)
	contractsGroup.Put("/:id", middleware.RequirePermission("contracts.update"), contractH.Update)
	contractsGroup.Delete("/:id", middleware.RequirePermission("contracts.delete"), contractH.Delete)

	// Meter Readings
	metersGroup := protected.Group("/meters")
	metersGroup.Get("/", middleware.RequirePermission("meters.read"), meterH.List)
	metersGroup.Post("/", middleware.RequirePermission("meters.create"), meterH.Create)
	metersGroup.Get("/:id", middleware.RequirePermission("meters.read"), meterH.GetByID)
	metersGroup.Put("/:id", middleware.RequirePermission("meters.update"), meterH.Update)
	metersGroup.Delete("/:id", middleware.RequirePermission("meters.delete"), meterH.Delete)
}
