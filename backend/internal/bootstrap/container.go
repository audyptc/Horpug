package bootstrap

import (
	"apigofiberhorpug/internal/database"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	authusecase "apigofiberhorpug/internal/feature/auth/usecase"
	billusecase "apigofiberhorpug/internal/feature/bill/usecase"
	menuusecase "apigofiberhorpug/internal/feature/menu/usecase"
	paymentusecase "apigofiberhorpug/internal/feature/payment/usecase"
	permissionusecase "apigofiberhorpug/internal/feature/permission/usecase"
	contractusecase "apigofiberhorpug/internal/feature/contract/usecase"
	electricmeterusecase "apigofiberhorpug/internal/feature/electricmeter/usecase"
	roleusecase "apigofiberhorpug/internal/feature/role/usecase"
	roomusecase "apigofiberhorpug/internal/feature/room/usecase"
	roomtypeusecase "apigofiberhorpug/internal/feature/roomtype/usecase"
	tenantusecase "apigofiberhorpug/internal/feature/tenant/usecase"
	userusecase "apigofiberhorpug/internal/feature/user/usecase"
	watermeterusecase "apigofiberhorpug/internal/feature/watermeter/usecase"
	"apigofiberhorpug/internal/usecase"
	"time"
)

type Container struct {
	AuthUC         *authusecase.AuthUseCase
	UserUC         *userusecase.UserUseCase
	RoleUC         *roleusecase.RoleUseCase
	PermUC         *permissionusecase.PermissionUseCase
	MenuUC         *menuusecase.MenuUseCase
	RoomUC         *roomusecase.RoomUseCase
	RoomTypeUC     *roomtypeusecase.RoomTypeUseCase
	TenantUC       *tenantusecase.TenantUseCase
	ContractUC     *contractusecase.ContractUseCase
	ElectricMeterUC *electricmeterusecase.ElectricMeterUseCase
	WaterMeterUC    *watermeterusecase.WaterMeterUseCase
	BillUC         *billusecase.BillUseCase
	DashboardUC    *usecase.DashboardUseCase
	AnalyticsUC    *usecase.AnalyticsUseCase
	ExpenseUC      *usecase.ExpenseUseCase
	MaintenanceUC  *usecase.MaintenanceRequestUseCase
	PaymentUC      *paymentusecase.PaymentUseCase
	AnnouncementUC *usecase.AnnouncementUseCase
	ReportUC       *usecase.ReportUseCase
	ParkingUC       *usecase.ParkingUseCase
	ParcelUC        *usecase.ParcelUseCase
	DocumentUC      *usecase.DocumentUseCase
	NotificationUC  *usecase.NotificationUseCase
	SearchUC        *usecase.SearchUseCase
	ActivityLogUC   *alusecase.ActivityLogUseCase
}

func NewContainer(db *database.DB, secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *Container {
	repos := newRepositories(db)
	return &Container{
		AuthUC:         authusecase.NewAuthUseCase(repos.user, repos.token, secretKey, accessTokenDuration, refreshTokenDuration),
		UserUC:         userusecase.NewUserUseCase(repos.user, repos.role),
		RoleUC:         roleusecase.NewRoleUseCase(repos.role, repos.perm, repos.menu),
		PermUC:         permissionusecase.NewPermissionUseCase(repos.perm),
		MenuUC:         menuusecase.NewMenuUseCase(repos.menu),
		RoomUC:         roomusecase.NewRoomUseCase(repos.room, repos.contract),
		RoomTypeUC:     roomtypeusecase.NewRoomTypeUseCase(repos.roomType),
		TenantUC:       tenantusecase.NewTenantUseCase(repos.tenant, repos.contract),
		ContractUC:     contractusecase.NewContractUseCase(repos.contract, repos.room, repos.tenant),
		ElectricMeterUC: electricmeterusecase.NewElectricMeterUseCase(repos.electricMeter, repos.room),
		WaterMeterUC:    watermeterusecase.NewWaterMeterUseCase(repos.waterMeter, repos.room),
		BillUC:         billusecase.NewBillUseCase(repos.bill, repos.contract),
		DashboardUC:    usecase.NewDashboardUseCase(repos.dashboard),
		AnalyticsUC:    usecase.NewAnalyticsUseCase(repos.analytics),
		ExpenseUC:      usecase.NewExpenseUseCase(repos.expense),
		MaintenanceUC:  usecase.NewMaintenanceRequestUseCase(repos.maintenance),
		PaymentUC:      paymentusecase.NewPaymentUseCase(repos.payment),
		AnnouncementUC: usecase.NewAnnouncementUseCase(repos.announcement),
		ReportUC:       usecase.NewReportUseCase(repos.report),
		ParkingUC:      usecase.NewParkingUseCase(repos.parking),
		ParcelUC:       usecase.NewParcelUseCase(repos.parcel),
		DocumentUC:     usecase.NewDocumentUseCase(repos.document),
		NotificationUC: usecase.NewNotificationUseCase(repos.notification),
		SearchUC:       usecase.NewSearchUseCase(repos.search),
		ActivityLogUC:  alusecase.NewActivityLogUseCase(repos.activityLog),
	}
}
