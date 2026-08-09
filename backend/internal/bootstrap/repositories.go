package bootstrap

import (
	"apigofiberhorpug/internal/database"
	alrepository "apigofiberhorpug/internal/feature/activitylog/repository"
	authrepository "apigofiberhorpug/internal/feature/auth/repository"
	billrepository "apigofiberhorpug/internal/feature/bill/repository"
	menurepository "apigofiberhorpug/internal/feature/menu/repository"
	paymentrepository "apigofiberhorpug/internal/feature/payment/repository"
	permissionrepository "apigofiberhorpug/internal/feature/permission/repository"
	contractrepository "apigofiberhorpug/internal/feature/contract/repository"
	electricmeterrepository "apigofiberhorpug/internal/feature/electricmeter/repository"
	rolerepository "apigofiberhorpug/internal/feature/role/repository"
	roomrepository "apigofiberhorpug/internal/feature/room/repository"
	roomtyperepository "apigofiberhorpug/internal/feature/roomtype/repository"
	tenantrepository "apigofiberhorpug/internal/feature/tenant/repository"
	userrepository "apigofiberhorpug/internal/feature/user/repository"
	watermeterrepository "apigofiberhorpug/internal/feature/watermeter/repository"
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
	contract     *contractrepository.ContractRepo
	electricMeter *electricmeterrepository.ElectricMeterRepo
	waterMeter    *watermeterrepository.WaterMeterRepo
	bill         *billrepository.BillRepo
	dashboard    *repository.DashboardRepo
	analytics    *repository.AnalyticsRepo
	expense      *repository.ExpenseRepo
	maintenance  *repository.MaintenanceRequestRepo
	payment      *paymentrepository.PaymentRepo
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
		contract:     contractrepository.NewContractRepo(db),
		electricMeter: electricmeterrepository.NewElectricMeterRepo(db),
		waterMeter:    watermeterrepository.NewWaterMeterRepo(db),
		bill:         billrepository.NewBillRepo(db),
		dashboard:    repository.NewDashboardRepo(db),
		analytics:    repository.NewAnalyticsRepo(db),
		expense:      repository.NewExpenseRepo(db),
		maintenance:  repository.NewMaintenanceRequestRepo(db),
		payment:      paymentrepository.NewPaymentRepo(db),
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
