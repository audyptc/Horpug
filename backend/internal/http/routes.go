package http

import (
	"time"

	activityloghttp "apihorpug/internal/features/activitylog/delivery/http"
	activitylogrepository "apihorpug/internal/features/activitylog/repository/postgres"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	announcementhttp "apihorpug/internal/features/announcement/delivery/http"
	announcementrepository "apihorpug/internal/features/announcement/repository/postgres"
	announcementusecase "apihorpug/internal/features/announcement/usecase"
	authhttp "apihorpug/internal/features/auth/delivery/http"
	authrepository "apihorpug/internal/features/auth/repository/postgres"
	authusecase "apihorpug/internal/features/auth/usecase"
	contracthttp "apihorpug/internal/features/contract/delivery/http"
	contractrepository "apihorpug/internal/features/contract/repository/postgres"
	contractusecase "apihorpug/internal/features/contract/usecase"
	documenthttp "apihorpug/internal/features/document/delivery/http"
	documentrepository "apihorpug/internal/features/document/repository/postgres"
	documentusecase "apihorpug/internal/features/document/usecase"
	dormitoryhttp "apihorpug/internal/features/dormitory/delivery/http"
	dormitoryrepository "apihorpug/internal/features/dormitory/repository/postgres"
	dormitoryusecase "apihorpug/internal/features/dormitory/usecase"
	expensehttp "apihorpug/internal/features/expense/delivery/http"
	expenserepository "apihorpug/internal/features/expense/repository/postgres"
	expenseusecase "apihorpug/internal/features/expense/usecase"
	invoicehttp "apihorpug/internal/features/invoice/delivery/http"
	invoicerepository "apihorpug/internal/features/invoice/repository/postgres"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"
	menuhttp "apihorpug/internal/features/menu/delivery/http"
	menurepository "apihorpug/internal/features/menu/repository/postgres"
	menuusecase "apihorpug/internal/features/menu/usecase"
	meterhttp "apihorpug/internal/features/meter/delivery/http"
	meterrepository "apihorpug/internal/features/meter/repository/postgres"
	meterusecase "apihorpug/internal/features/meter/usecase"
	parcelhttp "apihorpug/internal/features/parcel/delivery/http"
	parcelrepository "apihorpug/internal/features/parcel/repository/postgres"
	parcelusecase "apihorpug/internal/features/parcel/usecase"
	parkinghttp "apihorpug/internal/features/parking/delivery/http"
	parkingrepository "apihorpug/internal/features/parking/repository/postgres"
	parkingusecase "apihorpug/internal/features/parking/usecase"
	paymenthttp "apihorpug/internal/features/payment/delivery/http"
	paymentrepository "apihorpug/internal/features/payment/repository/postgres"
	paymentusecase "apihorpug/internal/features/payment/usecase"
	permissionhttp "apihorpug/internal/features/permission/delivery/http"
	permissiondomain "apihorpug/internal/features/permission/domain"
	permissionrepository "apihorpug/internal/features/permission/repository/postgres"
	permissionusecase "apihorpug/internal/features/permission/usecase"
	repairrequesthttp "apihorpug/internal/features/repairrequest/delivery/http"
	repairrequestrepository "apihorpug/internal/features/repairrequest/repository/postgres"
	repairrequestusecase "apihorpug/internal/features/repairrequest/usecase"
	rolehttp "apihorpug/internal/features/role/delivery/http"
	rolerepository "apihorpug/internal/features/role/repository/postgres"
	roleusecase "apihorpug/internal/features/role/usecase"
	roomhttp "apihorpug/internal/features/room/delivery/http"
	roomrepository "apihorpug/internal/features/room/repository/postgres"
	roomusecase "apihorpug/internal/features/room/usecase"
	roomtypehttp "apihorpug/internal/features/roomtype/delivery/http"
	roomtyperepository "apihorpug/internal/features/roomtype/repository/postgres"
	roomtypeusecase "apihorpug/internal/features/roomtype/usecase"
	tenanthttp "apihorpug/internal/features/tenant/delivery/http"
	tenantrepository "apihorpug/internal/features/tenant/repository/postgres"
	tenantusecase "apihorpug/internal/features/tenant/usecase"
	userhttp "apihorpug/internal/features/user/delivery/http"
	userrepository "apihorpug/internal/features/user/repository/postgres"
	userusecase "apihorpug/internal/features/user/usecase"
	watermeterhttp "apihorpug/internal/features/watermeter/delivery/http"
	watermeterrepository "apihorpug/internal/features/watermeter/repository/postgres"
	watermeterusecase "apihorpug/internal/features/watermeter/usecase"
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
	activityLogRepo := activitylogrepository.NewRepository(db)
	activityLogService := activitylogusecase.New(activityLogRepo)
	activityLogHandler := activityloghttp.NewHandler(activityLogService)
	userRepo := userrepository.NewRepository(db)
	userService := userusecase.New(userRepo, activityLogService)
	userHandler := userhttp.NewHandler(userService)
	roleRepo := rolerepository.NewRepository(db)
	roleService := roleusecase.New(roleRepo, activityLogService)
	roleHandler := rolehttp.NewHandler(roleService)
	dormitoryRepo := dormitoryrepository.NewRepository(db)
	dormitoryService := dormitoryusecase.New(dormitoryRepo, activityLogService)
	dormitoryHandler := dormitoryhttp.NewHandler(dormitoryService)
	roomTypeRepo := roomtyperepository.NewRepository(db)
	roomTypeService := roomtypeusecase.New(roomTypeRepo, activityLogService)
	roomTypeHandler := roomtypehttp.NewHandler(roomTypeService)
	roomRepo := roomrepository.NewRepository(db)
	roomService := roomusecase.New(roomRepo, activityLogService)
	roomHandler := roomhttp.NewHandler(roomService)
	tenantRepo := tenantrepository.NewRepository(db)
	tenantService := tenantusecase.New(tenantRepo)
	tenantHandler := tenanthttp.NewHandler(tenantService)
	contractRepo := contractrepository.NewRepository(db)
	contractService := contractusecase.New(contractRepo)
	contractHandler := contracthttp.NewHandler(contractService)
	meterRepo := meterrepository.NewRepository(db)
	meterService := meterusecase.New(meterRepo)
	meterHandler := meterhttp.NewHandler(meterService)
	waterMeterRepo := watermeterrepository.NewRepository(db)
	waterMeterService := watermeterusecase.New(waterMeterRepo)
	waterMeterHandler := watermeterhttp.NewHandler(waterMeterService)
	invoiceRepo := invoicerepository.NewRepository(db)
	invoiceService := invoiceusecase.New(invoiceRepo)
	invoiceHandler := invoicehttp.NewHandler(invoiceService)
	paymentRepo := paymentrepository.NewRepository(db)
	paymentService := paymentusecase.New(paymentRepo)
	paymentHandler := paymenthttp.NewHandler(paymentService)
	expenseRepo := expenserepository.NewRepository(db)
	expenseService := expenseusecase.New(expenseRepo)
	expenseHandler := expensehttp.NewHandler(expenseService)
	repairRequestRepo := repairrequestrepository.NewRepository(db)
	repairRequestService := repairrequestusecase.New(repairRequestRepo)
	repairRequestHandler := repairrequesthttp.NewHandler(repairRequestService)
	parkingRepo := parkingrepository.NewRepository(db)
	parkingService := parkingusecase.New(parkingRepo)
	parkingHandler := parkinghttp.NewHandler(parkingService)
	parcelRepo := parcelrepository.NewRepository(db)
	parcelService := parcelusecase.New(parcelRepo)
	parcelHandler := parcelhttp.NewHandler(parcelService)
	announcementRepo := announcementrepository.NewRepository(db)
	announcementService := announcementusecase.New(announcementRepo)
	announcementHandler := announcementhttp.NewHandler(announcementService)
	documentRepo := documentrepository.NewRepository(db)
	documentService := documentusecase.New(documentRepo)
	documentHandler := documenthttp.NewHandler(documentService)
	authTokenRepo := authrepository.NewRepository(db)
	authService := authusecase.New(userRepo, authTokenRepo, activityLogService, secretKey, accessTokenTTL, refreshTokenTTL)
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
	api.Get("/roles/active", requirePermission("/roles", permissiondomain.ActionRead), roleHandler.ListActive)
	api.Get("/roles/:id", requirePermission("/roles", permissiondomain.ActionRead), roleHandler.Get)
	api.Get("/roles/:id/deletion-check", requirePermission("/roles", permissiondomain.ActionDelete), roleHandler.CheckDeletion)
	api.Post("/roles", requirePermission("/roles", permissiondomain.ActionCreate), roleHandler.Create)
	api.Put("/roles/:id", requirePermission("/roles", permissiondomain.ActionUpdate), roleHandler.Update)
	api.Delete("/roles/:id", requirePermission("/roles", permissiondomain.ActionDelete), roleHandler.Delete)

	api.Get("/users", requirePermission("/users", permissiondomain.ActionRead), userHandler.List)
	api.Get("/users/active", requirePermission("/users", permissiondomain.ActionRead), userHandler.ListActive)
	api.Get("/users/:id", requirePermission("/users", permissiondomain.ActionRead), userHandler.Get)
	api.Get("/users/:id/permissions", requirePermission("/users", permissiondomain.ActionRead), userHandler.GetPermissions)
	api.Get("/users/:id/deletion-check", requirePermission("/users", permissiondomain.ActionDelete), userHandler.CheckDeletion)
	api.Post("/users", requirePermission("/users", permissiondomain.ActionCreate), userHandler.Create)
	api.Put("/users/:id", requirePermission("/users", permissiondomain.ActionUpdate), userHandler.Update)
	api.Delete("/users/:id", requirePermission("/users", permissiondomain.ActionDelete), userHandler.Delete)

	api.Get("/dormitories", requirePermission("/dormitories", permissiondomain.ActionRead), dormitoryHandler.List)
	api.Get("/dormitories/active", requirePermission("/dormitories", permissiondomain.ActionRead), dormitoryHandler.ListActive)
	api.Get("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionRead), dormitoryHandler.Get)
	api.Get("/dormitories/:id/deletion-check", requirePermission("/dormitories", permissiondomain.ActionDelete), dormitoryHandler.CheckDeletion)
	api.Post("/dormitories", requirePermission("/dormitories", permissiondomain.ActionCreate), dormitoryHandler.Create)
	api.Put("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionUpdate), dormitoryHandler.Update)
	api.Delete("/dormitories/:id", requirePermission("/dormitories", permissiondomain.ActionDelete), dormitoryHandler.Delete)

	api.Get("/room-types", requirePermission("/room-types", permissiondomain.ActionRead), roomTypeHandler.List)
	api.Get("/room-types/active", requirePermission("/room-types", permissiondomain.ActionRead), roomTypeHandler.ListActive)
	api.Get("/room-types/:id", requirePermission("/room-types", permissiondomain.ActionRead), roomTypeHandler.Get)
	api.Get("/room-types/:id/deletion-check", requirePermission("/room-types", permissiondomain.ActionDelete), roomTypeHandler.CheckDeletion)
	api.Post("/room-types", requirePermission("/room-types", permissiondomain.ActionCreate), roomTypeHandler.Create)
	api.Put("/room-types/:id", requirePermission("/room-types", permissiondomain.ActionUpdate), roomTypeHandler.Update)
	api.Delete("/room-types/:id", requirePermission("/room-types", permissiondomain.ActionDelete), roomTypeHandler.Delete)

	api.Get("/rooms", requirePermission("/rooms", permissiondomain.ActionRead), roomHandler.List)
	api.Get("/rooms/active", requirePermission("/rooms", permissiondomain.ActionRead), roomHandler.ListActive)
	api.Get("/rooms/:id", requirePermission("/rooms", permissiondomain.ActionRead), roomHandler.Get)
	api.Get("/rooms/:id/deletion-check", requirePermission("/rooms", permissiondomain.ActionDelete), roomHandler.CheckDeletion)
	api.Post("/rooms", requirePermission("/rooms", permissiondomain.ActionCreate), roomHandler.Create)
	api.Put("/rooms/:id", requirePermission("/rooms", permissiondomain.ActionUpdate), roomHandler.Update)
	api.Delete("/rooms/:id", requirePermission("/rooms", permissiondomain.ActionDelete), roomHandler.Delete)

	api.Get("/tenants", requirePermission("/tenants", permissiondomain.ActionRead), tenantHandler.List)
	api.Get("/tenants/active", requirePermission("/tenants", permissiondomain.ActionRead), tenantHandler.ListActive)
	api.Get("/tenants/:id", requirePermission("/tenants", permissiondomain.ActionRead), tenantHandler.Get)
	api.Post("/tenants", requirePermission("/tenants", permissiondomain.ActionCreate), tenantHandler.Create)
	api.Put("/tenants/:id", requirePermission("/tenants", permissiondomain.ActionUpdate), tenantHandler.Update)
	api.Delete("/tenants/:id", requirePermission("/tenants", permissiondomain.ActionDelete), tenantHandler.Delete)

	api.Get("/contracts", requirePermission("/contracts", permissiondomain.ActionRead), contractHandler.List)
	api.Get("/contracts/:id", requirePermission("/contracts", permissiondomain.ActionRead), contractHandler.Get)
	api.Post("/contracts", requirePermission("/contracts", permissiondomain.ActionCreate), contractHandler.Create)
	api.Put("/contracts/:id", requirePermission("/contracts", permissiondomain.ActionUpdate), contractHandler.Update)
	api.Delete("/contracts/:id", requirePermission("/contracts", permissiondomain.ActionDelete), contractHandler.Delete)

	api.Get("/meters", requirePermission("/meters", permissiondomain.ActionRead), meterHandler.List)
	api.Get("/meters/:id", requirePermission("/meters", permissiondomain.ActionRead), meterHandler.Get)
	api.Post("/meters", requirePermission("/meters", permissiondomain.ActionCreate), meterHandler.Create)
	api.Put("/meters/:id", requirePermission("/meters", permissiondomain.ActionUpdate), meterHandler.Update)
	api.Delete("/meters/:id", requirePermission("/meters", permissiondomain.ActionDelete), meterHandler.Delete)

	api.Get("/water-meters", requirePermission("/water-meters", permissiondomain.ActionRead), waterMeterHandler.List)
	api.Get("/water-meters/:id", requirePermission("/water-meters", permissiondomain.ActionRead), waterMeterHandler.Get)
	api.Post("/water-meters", requirePermission("/water-meters", permissiondomain.ActionCreate), waterMeterHandler.Create)
	api.Put("/water-meters/:id", requirePermission("/water-meters", permissiondomain.ActionUpdate), waterMeterHandler.Update)
	api.Delete("/water-meters/:id", requirePermission("/water-meters", permissiondomain.ActionDelete), waterMeterHandler.Delete)

	api.Get("/invoices", requirePermission("/invoices", permissiondomain.ActionRead), invoiceHandler.List)
	api.Get("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionRead), invoiceHandler.Get)
	api.Post("/invoices", requirePermission("/invoices", permissiondomain.ActionCreate), invoiceHandler.Create)
	api.Put("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionUpdate), invoiceHandler.Update)
	api.Delete("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionDelete), invoiceHandler.Delete)

	api.Get("/payments", requirePermission("/payments", permissiondomain.ActionRead), paymentHandler.List)
	api.Get("/payments/:id", requirePermission("/payments", permissiondomain.ActionRead), paymentHandler.Get)
	api.Post("/payments", requirePermission("/payments", permissiondomain.ActionCreate), paymentHandler.Create)
	api.Delete("/payments/:id", requirePermission("/payments", permissiondomain.ActionDelete), paymentHandler.Delete)

	api.Get("/expenses", requirePermission("/expenses", permissiondomain.ActionRead), expenseHandler.List)
	api.Get("/expenses/:id", requirePermission("/expenses", permissiondomain.ActionRead), expenseHandler.Get)
	api.Post("/expenses", requirePermission("/expenses", permissiondomain.ActionCreate), expenseHandler.Create)
	api.Put("/expenses/:id", requirePermission("/expenses", permissiondomain.ActionUpdate), expenseHandler.Update)
	api.Delete("/expenses/:id", requirePermission("/expenses", permissiondomain.ActionDelete), expenseHandler.Delete)

	api.Get("/repair-requests", requirePermission("/repair-requests", permissiondomain.ActionRead), repairRequestHandler.List)
	api.Get("/repair-requests/:id", requirePermission("/repair-requests", permissiondomain.ActionRead), repairRequestHandler.Get)
	api.Post("/repair-requests", requirePermission("/repair-requests", permissiondomain.ActionCreate), repairRequestHandler.Create)
	api.Put("/repair-requests/:id", requirePermission("/repair-requests", permissiondomain.ActionUpdate), repairRequestHandler.Update)
	api.Delete("/repair-requests/:id", requirePermission("/repair-requests", permissiondomain.ActionDelete), repairRequestHandler.Delete)

	api.Get("/parking", requirePermission("/parking", permissiondomain.ActionRead), parkingHandler.List)
	api.Get("/parking/:id", requirePermission("/parking", permissiondomain.ActionRead), parkingHandler.Get)
	api.Post("/parking", requirePermission("/parking", permissiondomain.ActionCreate), parkingHandler.Create)
	api.Put("/parking/:id", requirePermission("/parking", permissiondomain.ActionUpdate), parkingHandler.Update)
	api.Delete("/parking/:id", requirePermission("/parking", permissiondomain.ActionDelete), parkingHandler.Delete)

	api.Get("/parcels", requirePermission("/parcels", permissiondomain.ActionRead), parcelHandler.List)
	api.Get("/parcels/:id", requirePermission("/parcels", permissiondomain.ActionRead), parcelHandler.Get)
	api.Post("/parcels", requirePermission("/parcels", permissiondomain.ActionCreate), parcelHandler.Create)
	api.Put("/parcels/:id", requirePermission("/parcels", permissiondomain.ActionUpdate), parcelHandler.Update)
	api.Delete("/parcels/:id", requirePermission("/parcels", permissiondomain.ActionDelete), parcelHandler.Delete)

	api.Get("/activity-logs", requirePermission("/activity-logs", permissiondomain.ActionRead), activityLogHandler.List)
	api.Get("/activity-logs/:id", requirePermission("/activity-logs", permissiondomain.ActionRead), activityLogHandler.Get)
	api.Post("/activity-logs", requirePermission("/activity-logs", permissiondomain.ActionCreate), activityLogHandler.Create)

	api.Get("/announcements", requirePermission("/announcements", permissiondomain.ActionRead), announcementHandler.List)
	api.Get("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionRead), announcementHandler.Get)
	api.Post("/announcements", requirePermission("/announcements", permissiondomain.ActionCreate), announcementHandler.Create)
	api.Put("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionUpdate), announcementHandler.Update)
	api.Delete("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionDelete), announcementHandler.Delete)

	api.Get("/documents", requirePermission("/documents", permissiondomain.ActionRead), documentHandler.List)
	api.Get("/documents/:id", requirePermission("/documents", permissiondomain.ActionRead), documentHandler.Get)
	api.Post("/documents", requirePermission("/documents", permissiondomain.ActionCreate), documentHandler.Create)
	api.Put("/documents/:id", requirePermission("/documents", permissiondomain.ActionUpdate), documentHandler.Update)
	api.Delete("/documents/:id", requirePermission("/documents", permissiondomain.ActionDelete), documentHandler.Delete)
}
