package http

import (
	"time"

	"apigofiberhorpug/config"
	"apigofiberhorpug/internal/app/bootstrap"
	aldelivery "apigofiberhorpug/internal/feature/activitylog/delivery"
	analyticsdelivery "apigofiberhorpug/internal/feature/analytics/delivery"
	announcementdelivery "apigofiberhorpug/internal/feature/announcement/delivery"
	authdelivery "apigofiberhorpug/internal/feature/auth/delivery"
	billdelivery "apigofiberhorpug/internal/feature/bill/delivery"
	contractdelivery "apigofiberhorpug/internal/feature/contract/delivery"
	dashboarddelivery "apigofiberhorpug/internal/feature/dashboard/delivery"
	documentdelivery "apigofiberhorpug/internal/feature/document/delivery"
	dormitorydelivery "apigofiberhorpug/internal/feature/dormitory/delivery"
	electricmeterdelivery "apigofiberhorpug/internal/feature/electricmeter/delivery"
	expensedelivery "apigofiberhorpug/internal/feature/expense/delivery"
	maintenancerequestdelivery "apigofiberhorpug/internal/feature/maintenancerequest/delivery"
	menudelivery "apigofiberhorpug/internal/feature/menu/delivery"
	notificationdelivery "apigofiberhorpug/internal/feature/notification/delivery"
	parceldelivery "apigofiberhorpug/internal/feature/parcel/delivery"
	parkingdelivery "apigofiberhorpug/internal/feature/parking/delivery"
	paymentdelivery "apigofiberhorpug/internal/feature/payment/delivery"
	permissiondelivery "apigofiberhorpug/internal/feature/permission/delivery"
	reportdelivery "apigofiberhorpug/internal/feature/report/delivery"
	roledelivery "apigofiberhorpug/internal/feature/role/delivery"
	roomdelivery "apigofiberhorpug/internal/feature/room/delivery"
	roomtypedelivery "apigofiberhorpug/internal/feature/roomtype/delivery"
	searchdelivery "apigofiberhorpug/internal/feature/search/delivery"
	tenantdelivery "apigofiberhorpug/internal/feature/tenant/delivery"
	userdelivery "apigofiberhorpug/internal/feature/user/delivery"
	watermeterdelivery "apigofiberhorpug/internal/feature/watermeter/delivery"
	"apigofiberhorpug/internal/shared/http/apierror"
	"apigofiberhorpug/internal/shared/http/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func SetupRoutes(app *fiber.App, c *bootstrap.Container, cfg *config.Config) {
	authH := authdelivery.NewAuthHandler(c.AuthUC)
	userH := userdelivery.NewUserHandler(c.UserUC, c.ActivityLogUC)
	roleH := roledelivery.NewRoleHandler(c.RoleUC, c.ActivityLogUC)
	permH := permissiondelivery.NewPermissionHandler(c.PermUC)
	menuH := menudelivery.NewMenuHandler(c.MenuUC)
	roomH := roomdelivery.NewRoomHandler(c.RoomUC, c.ActivityLogUC)
	roomTypeH := roomtypedelivery.NewRoomTypeHandler(c.RoomTypeUC)
	tenantH := tenantdelivery.NewTenantHandler(c.TenantUC, c.ActivityLogUC)
	contractH := contractdelivery.NewContractHandler(c.ContractUC, c.ActivityLogUC)
	dormitoryH := dormitorydelivery.NewDormitoryHandler(c.DormitoryUC, c.ActivityLogUC)
	electricMeterH := electricmeterdelivery.NewElectricMeterHandler(c.ElectricMeterUC, c.ActivityLogUC)
	waterMeterH := watermeterdelivery.NewWaterMeterHandler(c.WaterMeterUC, c.ActivityLogUC)
	billH := billdelivery.NewBillHandler(c.BillUC, c.ActivityLogUC)
	dashboardH := dashboarddelivery.NewDashboardHandler(c.DashboardUC)
	analyticsH := analyticsdelivery.NewAnalyticsHandler(c.AnalyticsUC)
	expenseH := expensedelivery.NewExpenseHandler(c.ExpenseUC)
	maintenanceH := maintenancerequestdelivery.NewMaintenanceRequestHandler(c.MaintenanceUC)
	paymentH := paymentdelivery.NewPaymentHandler(c.PaymentUC, c.ActivityLogUC)
	announcementH := announcementdelivery.NewAnnouncementHandler(c.AnnouncementUC)
	reportH := reportdelivery.NewReportHandler(c.ReportUC)
	parkingH := parkingdelivery.NewParkingHandler(c.ParkingUC)
	parcelH := parceldelivery.NewParcelHandler(c.ParcelUC)
	documentH := documentdelivery.NewDocumentHandler(c.DocumentUC, cfg.UploadDir, cfg.UploadBaseURL)
	notificationH := notificationdelivery.NewNotificationHandler(c.NotificationUC)
	searchH := searchdelivery.NewSearchHandler(c.SearchUC)
	activityLogH := aldelivery.NewActivityLogHandler(c.ActivityLogUC)

	api := app.Group("/api/v1")

	loginLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return apierror.TooManyRequests("too many login attempts, please try again later")
		},
	})

	// Public: auth
	authGroup := api.Group("/auth")
	authGroup.Post("/login", loginLimiter, authH.Login)
	authGroup.Post("/refresh", authH.Refresh)
	authGroup.Post("/logout", authH.Logout)

	// Protected: all routes below require a valid JWT
	protected := api.Group("", middleware.RequireAuth(c.AuthUC))

	// Scoped: routes below additionally require a resolved dormitory (branch) context
	scoped := protected.Group("", middleware.RequireDormitory(c.DormitoryUC))

	// Dormitories (branches) — not itself dormitory-scoped
	dormitoriesGroup := protected.Group("/dormitories")
	dormitoriesGroup.Get("/", middleware.RequirePermission("settings/dormitories.read"), dormitoryH.List)
	dormitoriesGroup.Get("/mine", dormitoryH.Mine)
	dormitoriesGroup.Get("/users/:userId", middleware.RequirePermission("settings/dormitories.read"), dormitoryH.GetForUser)
	dormitoriesGroup.Post("/", middleware.RequirePermission("settings/dormitories.create"), dormitoryH.Create)
	dormitoriesGroup.Get("/:id", middleware.RequirePermission("settings/dormitories.read"), dormitoryH.GetByID)
	dormitoriesGroup.Put("/:id", middleware.RequirePermission("settings/dormitories.update"), dormitoryH.Update)
	dormitoriesGroup.Delete("/:id", middleware.RequirePermission("settings/dormitories.delete"), dormitoryH.Delete)
	dormitoriesGroup.Put("/users/:userId", middleware.RequirePermission("settings/dormitories.update"), dormitoryH.AssignToUser)

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

	// Permissions metadata
	permGroup := protected.Group("/permissions")
	permGroup.Get("/", middleware.RequirePermission("settings/roles.read"), permH.List)

	// Menus metadata
	menusGroup := protected.Group("/menus")
	menusGroup.Get("/", middleware.RequirePermission("settings/roles.read"), menuH.List)
	menusGroup.Get("/:id", middleware.RequirePermission("settings/roles.read"), menuH.GetByID)

	// Rooms
	roomsGroup := scoped.Group("/rooms")
	roomsGroup.Get("/", middleware.RequirePermission("rooms.read"), roomH.List)
	roomsGroup.Post("/", middleware.RequirePermission("rooms.create"), roomH.Create)
	roomsGroup.Get("/:id", middleware.RequirePermission("rooms.read"), roomH.GetByID)
	roomsGroup.Put("/:id", middleware.RequirePermission("rooms.update"), roomH.Update)
	roomsGroup.Delete("/:id", middleware.RequirePermission("rooms.delete"), roomH.Delete)

	// Room Types
	roomTypesGroup := protected.Group("/room-types")
	roomTypesGroup.Get("/", middleware.RequirePermission("settings/room-types.read"), roomTypeH.List)
	roomTypesGroup.Post("/", middleware.RequirePermission("settings/room-types.create"), roomTypeH.Create)
	roomTypesGroup.Get("/:id", middleware.RequirePermission("settings/room-types.read"), roomTypeH.GetByID)
	roomTypesGroup.Put("/:id", middleware.RequirePermission("settings/room-types.update"), roomTypeH.Update)
	roomTypesGroup.Delete("/:id", middleware.RequirePermission("settings/room-types.delete"), roomTypeH.Delete)

	// Tenants
	tenantsGroup := scoped.Group("/tenants")
	tenantsGroup.Get("/", middleware.RequirePermission("tenants.read"), tenantH.List)
	tenantsGroup.Post("/", middleware.RequirePermission("tenants.create"), tenantH.Create)
	tenantsGroup.Get("/:id", middleware.RequirePermission("tenants.read"), tenantH.GetByID)
	tenantsGroup.Put("/:id", middleware.RequirePermission("tenants.update"), tenantH.Update)
	tenantsGroup.Delete("/:id", middleware.RequirePermission("tenants.delete"), tenantH.Delete)

	// Contracts
	contractsGroup := scoped.Group("/contracts")
	contractsGroup.Get("/", middleware.RequirePermission("contracts.read"), contractH.List)
	contractsGroup.Post("/", middleware.RequirePermission("contracts.create"), contractH.Create)
	contractsGroup.Get("/:id", middleware.RequirePermission("contracts.read"), contractH.GetByID)
	contractsGroup.Put("/:id", middleware.RequirePermission("contracts.update"), contractH.Update)
	contractsGroup.Delete("/:id", middleware.RequirePermission("contracts.delete"), contractH.Delete)

	// Electric Meter Readings
	electricMetersGroup := scoped.Group("/electric-meters")
	electricMetersGroup.Get("/", middleware.RequirePermission("electric-meters.read"), electricMeterH.List)
	electricMetersGroup.Post("/", middleware.RequirePermission("electric-meters.create"), electricMeterH.Create)
	electricMetersGroup.Get("/latest", middleware.RequirePermission("electric-meters.read"), electricMeterH.GetLatestByRoomID)
	electricMetersGroup.Get("/:id", middleware.RequirePermission("electric-meters.read"), electricMeterH.GetByID)
	electricMetersGroup.Put("/:id", middleware.RequirePermission("electric-meters.update"), electricMeterH.Update)
	electricMetersGroup.Delete("/:id", middleware.RequirePermission("electric-meters.delete"), electricMeterH.Delete)

	// Water Meter Readings
	waterMetersGroup := scoped.Group("/water-meters")
	waterMetersGroup.Get("/", middleware.RequirePermission("water-meters.read"), waterMeterH.List)
	waterMetersGroup.Post("/", middleware.RequirePermission("water-meters.create"), waterMeterH.Create)
	waterMetersGroup.Get("/latest", middleware.RequirePermission("water-meters.read"), waterMeterH.GetLatestByRoomID)
	waterMetersGroup.Get("/:id", middleware.RequirePermission("water-meters.read"), waterMeterH.GetByID)
	waterMetersGroup.Put("/:id", middleware.RequirePermission("water-meters.update"), waterMeterH.Update)
	waterMetersGroup.Delete("/:id", middleware.RequirePermission("water-meters.delete"), waterMeterH.Delete)

	// Bills
	billsGroup := scoped.Group("/bills")
	billsGroup.Get("/", middleware.RequirePermission("bills.read"), billH.List)
	billsGroup.Post("/", middleware.RequirePermission("bills.create"), billH.Create)
	billsGroup.Get("/:id", middleware.RequirePermission("bills.read"), billH.GetByID)
	billsGroup.Put("/:id", middleware.RequirePermission("bills.update"), billH.Update)
	billsGroup.Delete("/:id", middleware.RequirePermission("bills.delete"), billH.Delete)

	// Dashboard
	scoped.Get("/dashboard/summary", middleware.RequirePermission(".read"), dashboardH.Summary)

	// Analytics
	scoped.Get("/analytics/summary", middleware.RequirePermission("analytics.read"), analyticsH.Summary)

	// Maintenance Requests
	maintenanceGroup := scoped.Group("/maintenance")
	maintenanceGroup.Get("/", middleware.RequirePermission("maintenance.read"), maintenanceH.List)
	maintenanceGroup.Post("/", middleware.RequirePermission("maintenance.create"), maintenanceH.Create)
	maintenanceGroup.Get("/:id", middleware.RequirePermission("maintenance.read"), maintenanceH.GetByID)
	maintenanceGroup.Put("/:id", middleware.RequirePermission("maintenance.update"), maintenanceH.Update)
	maintenanceGroup.Delete("/:id", middleware.RequirePermission("maintenance.delete"), maintenanceH.Delete)

	// Expenses
	expensesGroup := scoped.Group("/expenses")
	expensesGroup.Get("/", middleware.RequirePermission("expenses.read"), expenseH.List)
	expensesGroup.Post("/", middleware.RequirePermission("expenses.create"), expenseH.Create)
	expensesGroup.Get("/:id", middleware.RequirePermission("expenses.read"), expenseH.GetByID)
	expensesGroup.Put("/:id", middleware.RequirePermission("expenses.update"), expenseH.Update)
	expensesGroup.Delete("/:id", middleware.RequirePermission("expenses.delete"), expenseH.Delete)

	// Payments
	paymentsGroup := scoped.Group("/payments")
	paymentsGroup.Get("/", middleware.RequirePermission("payments.read"), paymentH.List)
	paymentsGroup.Post("/", middleware.RequirePermission("payments.create"), paymentH.Create)
	paymentsGroup.Get("/:id", middleware.RequirePermission("payments.read"), paymentH.GetByID)
	paymentsGroup.Put("/:id", middleware.RequirePermission("payments.update"), paymentH.Update)
	paymentsGroup.Delete("/:id", middleware.RequirePermission("payments.delete"), paymentH.Delete)

	// Announcements
	announcementsGroup := scoped.Group("/announcements")
	announcementsGroup.Get("/", middleware.RequirePermission("announcements.read"), announcementH.List)
	announcementsGroup.Post("/", middleware.RequirePermission("announcements.create"), announcementH.Create)
	announcementsGroup.Get("/:id", middleware.RequirePermission("announcements.read"), announcementH.GetByID)
	announcementsGroup.Put("/:id", middleware.RequirePermission("announcements.update"), announcementH.Update)
	announcementsGroup.Delete("/:id", middleware.RequirePermission("announcements.delete"), announcementH.Delete)

	// Reports
	reportsGroup := scoped.Group("/reports")
	reportsGroup.Get("/income", middleware.RequirePermission("reports.read"), reportH.Income)
	reportsGroup.Get("/expenses", middleware.RequirePermission("reports.read"), reportH.Expenses)
	reportsGroup.Get("/occupancy", middleware.RequirePermission("reports.read"), reportH.Occupancy)

	// Parking
	parkingGroup := scoped.Group("/parking")
	parkingGroup.Get("/", middleware.RequirePermission("parking.read"), parkingH.List)
	parkingGroup.Post("/", middleware.RequirePermission("parking.create"), parkingH.Create)
	parkingGroup.Get("/:id", middleware.RequirePermission("parking.read"), parkingH.GetByID)
	parkingGroup.Put("/:id", middleware.RequirePermission("parking.update"), parkingH.Update)
	parkingGroup.Delete("/:id", middleware.RequirePermission("parking.delete"), parkingH.Delete)

	// Parcels
	parcelsGroup := scoped.Group("/parcels")
	parcelsGroup.Get("/", middleware.RequirePermission("parcels.read"), parcelH.List)
	parcelsGroup.Post("/", middleware.RequirePermission("parcels.create"), parcelH.Create)
	parcelsGroup.Get("/:id", middleware.RequirePermission("parcels.read"), parcelH.GetByID)
	parcelsGroup.Put("/:id", middleware.RequirePermission("parcels.update"), parcelH.Update)
	parcelsGroup.Delete("/:id", middleware.RequirePermission("parcels.delete"), parcelH.Delete)

	// Documents
	documentsGroup := scoped.Group("/documents")
	documentsGroup.Get("/", middleware.RequirePermission("documents.read"), documentH.List)
	documentsGroup.Post("/", middleware.RequirePermission("documents.create"), documentH.Create)
	documentsGroup.Post("/upload", middleware.RequirePermission("documents.create"), documentH.Upload)
	documentsGroup.Get("/:id", middleware.RequirePermission("documents.read"), documentH.GetByID)
	documentsGroup.Put("/:id", middleware.RequirePermission("documents.update"), documentH.Update)
	documentsGroup.Delete("/:id", middleware.RequirePermission("documents.delete"), documentH.Delete)

	// Notifications
	scoped.Get("/notifications", middleware.RequirePermission("notifications.read"), notificationH.List)

	// Search
	scoped.Get("/search", middleware.RequirePermission("search.read"), searchH.Global)

	// Activity Logs
	scoped.Get("/activity-logs", middleware.RequirePermission("activity-logs.read"), activityLogH.List)
}
