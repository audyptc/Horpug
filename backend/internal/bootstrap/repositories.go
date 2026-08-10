package bootstrap

import (
	"apigofiberhorpug/internal/database"
	alrepository "apigofiberhorpug/internal/feature/activitylog/repository"
	announcementrepository "apigofiberhorpug/internal/feature/announcement/repository"
	authrepository "apigofiberhorpug/internal/feature/auth/repository"
	billrepository "apigofiberhorpug/internal/feature/bill/repository"
	contractrepository "apigofiberhorpug/internal/feature/contract/repository"
	documentrepository "apigofiberhorpug/internal/feature/document/repository"
	electricmeterrepository "apigofiberhorpug/internal/feature/electricmeter/repository"
	expenserepository "apigofiberhorpug/internal/feature/expense/repository"
	maintenancerequestrepository "apigofiberhorpug/internal/feature/maintenancerequest/repository"
	menurepository "apigofiberhorpug/internal/feature/menu/repository"
	notificationrepository "apigofiberhorpug/internal/feature/notification/repository"
	parcelrepository "apigofiberhorpug/internal/feature/parcel/repository"
	parkingrepository "apigofiberhorpug/internal/feature/parking/repository"
	paymentrepository "apigofiberhorpug/internal/feature/payment/repository"
	permissionrepository "apigofiberhorpug/internal/feature/permission/repository"
	rolerepository "apigofiberhorpug/internal/feature/role/repository"
	roomrepository "apigofiberhorpug/internal/feature/room/repository"
	roomtyperepository "apigofiberhorpug/internal/feature/roomtype/repository"
	tenantrepository "apigofiberhorpug/internal/feature/tenant/repository"
	userrepository "apigofiberhorpug/internal/feature/user/repository"
	watermeterrepository "apigofiberhorpug/internal/feature/watermeter/repository"
	"apigofiberhorpug/internal/repository"
)

type repositories struct {
	user          *userrepository.UserRepo
	role          *rolerepository.RoleRepo
	perm          *permissionrepository.PermissionRepo
	token         *authrepository.TokenRepo
	menu          *menurepository.MenuRepo
	room          *roomrepository.RoomRepo
	roomType      *roomtyperepository.RoomTypeRepo
	tenant        *tenantrepository.TenantRepo
	contract      *contractrepository.ContractRepo
	electricMeter *electricmeterrepository.ElectricMeterRepo
	waterMeter    *watermeterrepository.WaterMeterRepo
	bill          *billrepository.BillRepo
	dashboard     *repository.DashboardRepo
	analytics     *repository.AnalyticsRepo
	expense       *expenserepository.ExpenseRepo
	maintenance   *maintenancerequestrepository.MaintenanceRequestRepo
	payment       *paymentrepository.PaymentRepo
	announcement  *announcementrepository.AnnouncementRepo
	report        *repository.ReportRepo
	parking       *parkingrepository.ParkingRepo
	parcel        *parcelrepository.ParcelRepo
	document      *documentrepository.DocumentRepo
	notification  *notificationrepository.NotificationRepo
	search        *repository.SearchRepo
	activityLog   *alrepository.ActivityLogRepo
}

func newRepositories(db *database.DB) repositories {
	return repositories{
		user:          userrepository.NewUserRepo(db),
		role:          rolerepository.NewRoleRepo(db),
		perm:          permissionrepository.NewPermissionRepo(db),
		token:         authrepository.NewTokenRepo(db),
		menu:          menurepository.NewMenuRepo(db),
		room:          roomrepository.NewRoomRepo(db),
		roomType:      roomtyperepository.NewRoomTypeRepo(db),
		tenant:        tenantrepository.NewTenantRepo(db),
		contract:      contractrepository.NewContractRepo(db),
		electricMeter: electricmeterrepository.NewElectricMeterRepo(db),
		waterMeter:    watermeterrepository.NewWaterMeterRepo(db),
		bill:          billrepository.NewBillRepo(db),
		dashboard:     repository.NewDashboardRepo(db),
		analytics:     repository.NewAnalyticsRepo(db),
		expense:       expenserepository.NewExpenseRepo(db),
		maintenance:   maintenancerequestrepository.NewMaintenanceRequestRepo(db),
		payment:       paymentrepository.NewPaymentRepo(db),
		announcement:  announcementrepository.NewAnnouncementRepo(db),
		report:        repository.NewReportRepo(db),
		parking:       parkingrepository.NewParkingRepo(db),
		parcel:        parcelrepository.NewParcelRepo(db),
		document:      documentrepository.NewDocumentRepo(db),
		notification:  notificationrepository.NewNotificationRepo(db),
		search:        repository.NewSearchRepo(db),
		activityLog:   alrepository.NewActivityLogRepo(db),
	}
}
