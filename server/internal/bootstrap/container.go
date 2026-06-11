package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/usecase"
	"time"
)

type Container struct {
	AuthUC         *usecase.AuthUseCase
	UserUC         *usecase.UserUseCase
	RoleUC         *usecase.RoleUseCase
	PermUC         *usecase.PermissionUseCase
	MenuUC         *usecase.MenuUseCase
	RoomUC         *usecase.RoomUseCase
	RoomTypeUC     *usecase.RoomTypeUseCase
	TenantUC       *usecase.TenantUseCase
	ContractUC     *usecase.ContractUseCase
	MeterReadingUC *usecase.MeterReadingUseCase
	BillUC         *usecase.BillUseCase
	DashboardUC    *usecase.DashboardUseCase
	AnalyticsUC    *usecase.AnalyticsUseCase
	ExpenseUC      *usecase.ExpenseUseCase
	MaintenanceUC  *usecase.MaintenanceRequestUseCase
	PaymentUC      *usecase.PaymentUseCase
	AnnouncementUC *usecase.AnnouncementUseCase
	ReportUC       *usecase.ReportUseCase
	ParkingUC       *usecase.ParkingUseCase
	ParcelUC        *usecase.ParcelUseCase
	DocumentUC      *usecase.DocumentUseCase
	NotificationUC  *usecase.NotificationUseCase
	SearchUC        *usecase.SearchUseCase
	ActivityLogUC   *usecase.ActivityLogUseCase
}

func NewContainer(db *database.DB, secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *Container {
	repos := newRepositories(db)
	return &Container{
		AuthUC:         usecase.NewAuthUseCase(repos.user, repos.token, secretKey, accessTokenDuration, refreshTokenDuration),
		UserUC:         usecase.NewUserUseCase(repos.user, repos.role),
		RoleUC:         usecase.NewRoleUseCase(repos.role, repos.perm, repos.menu),
		PermUC:         usecase.NewPermissionUseCase(repos.perm),
		MenuUC:         usecase.NewMenuUseCase(repos.menu),
		RoomUC:         usecase.NewRoomUseCase(repos.room, repos.contract),
		RoomTypeUC:     usecase.NewRoomTypeUseCase(repos.roomType),
		TenantUC:       usecase.NewTenantUseCase(repos.tenant),
		ContractUC:     usecase.NewContractUseCase(repos.contract, repos.room, repos.tenant),
		MeterReadingUC: usecase.NewMeterReadingUseCase(repos.meterReading, repos.room),
		BillUC:         usecase.NewBillUseCase(repos.bill, repos.contract),
		DashboardUC:    usecase.NewDashboardUseCase(repos.dashboard),
		AnalyticsUC:    usecase.NewAnalyticsUseCase(repos.analytics),
		ExpenseUC:      usecase.NewExpenseUseCase(repos.expense),
		MaintenanceUC:  usecase.NewMaintenanceRequestUseCase(repos.maintenance),
		PaymentUC:      usecase.NewPaymentUseCase(repos.payment),
		AnnouncementUC: usecase.NewAnnouncementUseCase(repos.announcement),
		ReportUC:       usecase.NewReportUseCase(repos.report),
		ParkingUC:      usecase.NewParkingUseCase(repos.parking),
		ParcelUC:       usecase.NewParcelUseCase(repos.parcel),
		DocumentUC:     usecase.NewDocumentUseCase(repos.document),
		NotificationUC: usecase.NewNotificationUseCase(repos.notification),
		SearchUC:       usecase.NewSearchUseCase(repos.search),
		ActivityLogUC:  usecase.NewActivityLogUseCase(repos.activityLog),
	}
}
