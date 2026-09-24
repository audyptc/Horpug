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
	dashboardhttp "apihorpug/internal/features/dashboard/delivery/http"
	dashboardrepository "apihorpug/internal/features/dashboard/repository/postgres"
	dashboardusecase "apihorpug/internal/features/dashboard/usecase"
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
	"apihorpug/internal/http/apierror"
	"apihorpug/internal/http/middleware"
	"apihorpug/internal/platform/filestore"
	"apihorpug/internal/platform/lineapi"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool, secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, cookieSecure bool, lineChannelAccessToken, lineChannelID string, files *filestore.Local) {
	lineClient := lineapi.New(lineChannelAccessToken, lineChannelID)

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
	tenantService := tenantusecase.New(tenantRepo, activityLogService, lineClient)
	tenantHandler := tenanthttp.NewHandler(tenantService)
	contractRepo := contractrepository.NewRepository(db)
	contractService := contractusecase.New(contractRepo, activityLogService, roomService)
	contractHandler := contracthttp.NewHandler(contractService)
	meterRepo := meterrepository.NewRepository(db)
	meterService := meterusecase.New(meterRepo, activityLogService)
	meterHandler := meterhttp.NewHandler(meterService)
	waterMeterRepo := watermeterrepository.NewRepository(db)
	waterMeterService := watermeterusecase.New(waterMeterRepo, activityLogService)
	waterMeterHandler := watermeterhttp.NewHandler(waterMeterService)
	invoiceRepo := invoicerepository.NewRepository(db)
	invoiceService := invoiceusecase.New(invoiceRepo, lineClient, activityLogService)
	invoiceHandler := invoicehttp.NewHandler(invoiceService)
	paymentRepo := paymentrepository.NewRepository(db)
	paymentService := paymentusecase.New(paymentRepo, activityLogService)
	paymentHandler := paymenthttp.NewHandler(paymentService)
	expenseRepo := expenserepository.NewRepository(db)
	expenseService := expenseusecase.New(expenseRepo, activityLogService)
	expenseHandler := expensehttp.NewHandler(expenseService)
	repairRequestRepo := repairrequestrepository.NewRepository(db)
	repairRequestService := repairrequestusecase.New(repairRequestRepo, activityLogService)
	repairRequestHandler := repairrequesthttp.NewHandler(repairRequestService)
	parkingRepo := parkingrepository.NewRepository(db)
	parkingService := parkingusecase.New(parkingRepo, activityLogService)
	parkingHandler := parkinghttp.NewHandler(parkingService)
	parcelRepo := parcelrepository.NewRepository(db)
	parcelService := parcelusecase.New(parcelRepo, activityLogService)
	parcelHandler := parcelhttp.NewHandler(parcelService)
	announcementRepo := announcementrepository.NewRepository(db)
	announcementService := announcementusecase.New(announcementRepo, activityLogService)
	announcementHandler := announcementhttp.NewHandler(announcementService)
	documentRepo := documentrepository.NewRepository(db)
	documentService := documentusecase.New(documentRepo, files, activityLogService)
	documentHandler := documenthttp.NewHandler(documentService)
	dashboardRepo := dashboardrepository.NewRepository(db)
	dashboardService := dashboardusecase.New(dashboardRepo)
	dashboardHandler := dashboardhttp.NewHandler(dashboardService)
	authTokenRepo := authrepository.NewRepository(db)
	authService := authusecase.New(userRepo, authTokenRepo, activityLogService, secretKey, accessTokenTTL, refreshTokenTTL)
	authHandler := authhttp.NewHandler(authService, cookieSecure)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/login", rateLimit(10, 5*time.Minute), authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	// Unauthenticated: called from the LIFF linking page a tenant opens from
	// their own LINE app, before they have any session with this system.
	publicGroup := app.Group("/api/v1/public")
	publicGroup.Post("/tenants/:id/line/link", rateLimit(20, 10*time.Minute), tenantHandler.LinkLine)
	publicGroup.Get("/line/oa", tenantHandler.LineOAInfo)

	api := app.Group("/api/v1")
	api.Use(middleware.RequireAuth(secretKey))

	requirePermission := func(menuPath string, action permissiondomain.Action) fiber.Handler {
		return middleware.RequirePermission(db, menuPath, action)
	}

	api.Get("/permissions", requirePermission("/permissions", permissiondomain.ActionRead), permissionHandler.List)
	api.Post("/permissions", requirePermission("/permissions", permissiondomain.ActionCreate), permissionHandler.Create)

	// The dashboard isn't a menu, so there is no route-level permission: the
	// handler gates each section on its own menu's read permission instead.
	api.Get("/dashboard/summary", dashboardHandler.GetSummary)

	api.Get("/menus", menuHandler.List)
	api.Get("/menus/mine", menuHandler.ListMine)

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
	api.Delete("/tenants/:id/line", requirePermission("/tenants", permissiondomain.ActionUpdate), tenantHandler.UnlinkLine)
	api.Get("/tenants/:id/line/status", requirePermission("/tenants", permissiondomain.ActionRead), tenantHandler.LineStatus)
	api.Get("/tenants/:id/deletion-check", requirePermission("/tenants", permissiondomain.ActionDelete), tenantHandler.CheckDeletion)
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
	// Registered before /:id so "generate" isn't taken for an invoice id.
	api.Get("/invoices/generate/preview", requirePermission("/invoices", permissiondomain.ActionCreate), invoiceHandler.PreviewGeneration)
	api.Post("/invoices/generate", requirePermission("/invoices", permissiondomain.ActionCreate), invoiceHandler.Generate)
	api.Get("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionRead), invoiceHandler.Get)
	api.Post("/invoices", requirePermission("/invoices", permissiondomain.ActionCreate), invoiceHandler.Create)
	api.Put("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionUpdate), invoiceHandler.Update)
	api.Delete("/invoices/:id", requirePermission("/invoices", permissiondomain.ActionDelete), invoiceHandler.Delete)
	api.Post("/invoices/:id/items", requirePermission("/invoices", permissiondomain.ActionUpdate), invoiceHandler.AddItem)
	api.Delete("/invoices/:id/items/:itemId", requirePermission("/invoices", permissiondomain.ActionUpdate), invoiceHandler.RemoveItem)
	api.Post("/invoices/:id/send-line", requirePermission("/invoices", permissiondomain.ActionUpdate), invoiceHandler.SendLine)
	api.Get("/invoices/:id/document", requirePermission("/invoices", permissiondomain.ActionRead), invoiceHandler.Document)
	api.Get("/invoices/:id/line-message", requirePermission("/invoices", permissiondomain.ActionRead), invoiceHandler.LineMessagePreview)

	api.Get("/payments", requirePermission("/payments", permissiondomain.ActionRead), paymentHandler.List)
	api.Get("/payments/:id", requirePermission("/payments", permissiondomain.ActionRead), paymentHandler.Get)
	api.Post("/payments", requirePermission("/payments", permissiondomain.ActionCreate), paymentHandler.Create)
	api.Put("/payments/:id", requirePermission("/payments", permissiondomain.ActionUpdate), paymentHandler.Update)
	api.Get("/payments/:id/receipt", requirePermission("/payments", permissiondomain.ActionRead), paymentHandler.Receipt)
	// Payments carry receipt numbers, so they are voided, never deleted.
	api.Post("/payments/:id/void", requirePermission("/payments", permissiondomain.ActionDelete), paymentHandler.Void)

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
	// Registered before /:id so "summary" isn't taken for an announcement id.
	api.Get("/announcements/summary", requirePermission("/announcements", permissiondomain.ActionRead), announcementHandler.Summary)
	api.Get("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionRead), announcementHandler.Get)
	api.Post("/announcements/:id/read", requirePermission("/announcements", permissiondomain.ActionRead), announcementHandler.MarkRead)
	api.Post("/announcements", requirePermission("/announcements", permissiondomain.ActionCreate), announcementHandler.Create)
	api.Put("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionUpdate), announcementHandler.Update)
	api.Delete("/announcements/:id", requirePermission("/announcements", permissiondomain.ActionDelete), announcementHandler.Delete)

	api.Get("/documents", requirePermission("/documents", permissiondomain.ActionRead), documentHandler.List)
	api.Get("/documents/:id", requirePermission("/documents", permissiondomain.ActionRead), documentHandler.Get)
	api.Get("/documents/:id/file", requirePermission("/documents", permissiondomain.ActionRead), documentHandler.File)
	api.Post("/documents", requirePermission("/documents", permissiondomain.ActionCreate), documentHandler.Create)
	api.Put("/documents/:id", requirePermission("/documents", permissiondomain.ActionUpdate), documentHandler.Update)
	api.Delete("/documents/:id", requirePermission("/documents", permissiondomain.ActionDelete), documentHandler.Delete)
}

// rateLimit caps requests per client IP (the real one, recovered from nginx's
// X-Forwarded-For — see main.go) on endpoints anyone can hit without a
// session, so passwords and tenant ids can't be brute-forced. Counts live in
// memory, which is enough for the single backend instance.
func rateLimit(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		LimitReached: func(c fiber.Ctx) error {
			return apierror.TooManyRequests("too many attempts, please try again later").WithSlug("too_many_requests")
		},
	})
}
