package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/repository"
)

type repositories struct {
	user         *repository.UserRepo
	role         *repository.RoleRepo
	perm         *repository.PermissionRepo
	token        *repository.TokenRepo
	menu         *repository.MenuRepo
	room         *repository.RoomRepo
	tenant       *repository.TenantRepo
	contract     *repository.ContractRepo
	meterReading *repository.MeterReadingRepo
	bill         *repository.BillRepo
	dashboard    *repository.DashboardRepo
	analytics    *repository.AnalyticsRepo
	expense      *repository.ExpenseRepo
	maintenance  *repository.MaintenanceRequestRepo
	payment      *repository.PaymentRepo
	announcement *repository.AnnouncementRepo
	report       *repository.ReportRepo
	parking      *repository.ParkingSlotRepo
}

func newRepositories(db *database.DB) repositories {
	return repositories{
		user:         repository.NewUserRepo(db),
		role:         repository.NewRoleRepo(db),
		perm:         repository.NewPermissionRepo(db),
		token:        repository.NewTokenRepo(db),
		menu:         repository.NewMenuRepo(db),
		room:         repository.NewRoomRepo(db),
		tenant:       repository.NewTenantRepo(db),
		contract:     repository.NewContractRepo(db),
		meterReading: repository.NewMeterReadingRepo(db),
		bill:         repository.NewBillRepo(db),
		dashboard:    repository.NewDashboardRepo(db),
		analytics:    repository.NewAnalyticsRepo(db),
		expense:      repository.NewExpenseRepo(db),
		maintenance:  repository.NewMaintenanceRequestRepo(db),
		payment:      repository.NewPaymentRepo(db),
		announcement: repository.NewAnnouncementRepo(db),
		report:       repository.NewReportRepo(db),
		parking:      repository.NewParkingSlotRepo(db),
	}
}
