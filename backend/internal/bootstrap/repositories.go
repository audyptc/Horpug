package bootstrap

import (
	"apigofiberhorpug/internal/database"
	alrepository "apigofiberhorpug/internal/feature/activitylog/repository"
	authrepository "apigofiberhorpug/internal/feature/auth/repository"
	menurepository "apigofiberhorpug/internal/feature/menu/repository"
	permissionrepository "apigofiberhorpug/internal/feature/permission/repository"
	rolerepository "apigofiberhorpug/internal/feature/role/repository"
	roomrepository "apigofiberhorpug/internal/feature/room/repository"
	roomtyperepository "apigofiberhorpug/internal/feature/roomtype/repository"
	tenantrepository "apigofiberhorpug/internal/feature/tenant/repository"
	userrepository "apigofiberhorpug/internal/feature/user/repository"
	"apigofiberhorpug/internal/repository"
)

type repositories struct {
	user         *userrepository.UserRepo
	role         *rolerepository.RoleRepo
	perm         *permissionrepository.PermissionRepo
	token        *authrepository.TokenRepo
	menu         *menurepository.MenuRepo
	room         *roomrepository.RoomRepo
	roomType     *roomtyperepository.RoomTypeRepo
	tenant       *tenantrepository.TenantRepo
	contract     *repository.ContractRepo
	electricMeter *repository.ElectricMeterRepo
	waterMeter    *repository.WaterMeterRepo
	bill         *repository.BillRepo
	dashboard    *repository.DashboardRepo
	analytics    *repository.AnalyticsRepo
	expense      *repository.ExpenseRepo
	maintenance  *repository.MaintenanceRequestRepo
	payment      *repository.PaymentRepo
	announcement *repository.AnnouncementRepo
	report       *repository.ReportRepo
	parking      *repository.ParkingRepo
	parcel       *repository.ParcelRepo
	document     *repository.DocumentRepo
	notification *repository.NotificationRepo
	search       *repository.SearchRepo
	activityLog  *alrepository.ActivityLogRepo
}

func newRepositories(db *database.DB) repositories {
	return repositories{
		user:         userrepository.NewUserRepo(db),
		role:         rolerepository.NewRoleRepo(db),
		perm:         permissionrepository.NewPermissionRepo(db),
		token:        authrepository.NewTokenRepo(db),
		menu:         menurepository.NewMenuRepo(db),
		room:         roomrepository.NewRoomRepo(db),
		roomType:     roomtyperepository.NewRoomTypeRepo(db),
		tenant:       tenantrepository.NewTenantRepo(db),
		contract:     repository.NewContractRepo(db),
		electricMeter: repository.NewElectricMeterRepo(db),
		waterMeter:    repository.NewWaterMeterRepo(db),
		bill:         repository.NewBillRepo(db),
		dashboard:    repository.NewDashboardRepo(db),
		analytics:    repository.NewAnalyticsRepo(db),
		expense:      repository.NewExpenseRepo(db),
		maintenance:  repository.NewMaintenanceRequestRepo(db),
		payment:      repository.NewPaymentRepo(db),
		announcement: repository.NewAnnouncementRepo(db),
		report:       repository.NewReportRepo(db),
		parking:      repository.NewParkingRepo(db),
		parcel:       repository.NewParcelRepo(db),
		document:     repository.NewDocumentRepo(db),
		notification: repository.NewNotificationRepo(db),
		search:       repository.NewSearchRepo(db),
		activityLog:  alrepository.NewActivityLogRepo(db),
	}
}
