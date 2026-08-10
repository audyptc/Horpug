package bootstrap

import (
	"apigofiberhorpug/internal/database"
	alusecase "apigofiberhorpug/internal/feature/activitylog/usecase"
	analyticsusecase "apigofiberhorpug/internal/feature/analytics/usecase"
	announcementusecase "apigofiberhorpug/internal/feature/announcement/usecase"
	authusecase "apigofiberhorpug/internal/feature/auth/usecase"
	billusecase "apigofiberhorpug/internal/feature/bill/usecase"
	contractusecase "apigofiberhorpug/internal/feature/contract/usecase"
	dashboardusecase "apigofiberhorpug/internal/feature/dashboard/usecase"
	documentusecase "apigofiberhorpug/internal/feature/document/usecase"
	dormitoryusecase "apigofiberhorpug/internal/feature/dormitory/usecase"
	electricmeterusecase "apigofiberhorpug/internal/feature/electricmeter/usecase"
	expenseusecase "apigofiberhorpug/internal/feature/expense/usecase"
	maintenancerequestusecase "apigofiberhorpug/internal/feature/maintenancerequest/usecase"
	menuusecase "apigofiberhorpug/internal/feature/menu/usecase"
	notificationusecase "apigofiberhorpug/internal/feature/notification/usecase"
	parcelusecase "apigofiberhorpug/internal/feature/parcel/usecase"
	parkingusecase "apigofiberhorpug/internal/feature/parking/usecase"
	paymentusecase "apigofiberhorpug/internal/feature/payment/usecase"
	permissionusecase "apigofiberhorpug/internal/feature/permission/usecase"
	reportusecase "apigofiberhorpug/internal/feature/report/usecase"
	roleusecase "apigofiberhorpug/internal/feature/role/usecase"
	roomusecase "apigofiberhorpug/internal/feature/room/usecase"
	roomtypeusecase "apigofiberhorpug/internal/feature/roomtype/usecase"
	searchusecase "apigofiberhorpug/internal/feature/search/usecase"
	tenantusecase "apigofiberhorpug/internal/feature/tenant/usecase"
	userusecase "apigofiberhorpug/internal/feature/user/usecase"
	watermeterusecase "apigofiberhorpug/internal/feature/watermeter/usecase"
	"time"
)

type Container struct {
	AuthUC          *authusecase.AuthUseCase
	UserUC          *userusecase.UserUseCase
	RoleUC          *roleusecase.RoleUseCase
	PermUC          *permissionusecase.PermissionUseCase
	MenuUC          *menuusecase.MenuUseCase
	RoomUC          *roomusecase.RoomUseCase
	RoomTypeUC      *roomtypeusecase.RoomTypeUseCase
	TenantUC        *tenantusecase.TenantUseCase
	ContractUC      *contractusecase.ContractUseCase
	DormitoryUC     *dormitoryusecase.DormitoryUseCase
	ElectricMeterUC *electricmeterusecase.ElectricMeterUseCase
	WaterMeterUC    *watermeterusecase.WaterMeterUseCase
	BillUC          *billusecase.BillUseCase
	DashboardUC     *dashboardusecase.DashboardUseCase
	AnalyticsUC     *analyticsusecase.AnalyticsUseCase
	ExpenseUC       *expenseusecase.ExpenseUseCase
	MaintenanceUC   *maintenancerequestusecase.MaintenanceRequestUseCase
	PaymentUC       *paymentusecase.PaymentUseCase
	AnnouncementUC  *announcementusecase.AnnouncementUseCase
	ReportUC        *reportusecase.ReportUseCase
	ParkingUC       *parkingusecase.ParkingUseCase
	ParcelUC        *parcelusecase.ParcelUseCase
	DocumentUC      *documentusecase.DocumentUseCase
	NotificationUC  *notificationusecase.NotificationUseCase
	SearchUC        *searchusecase.SearchUseCase
	ActivityLogUC   *alusecase.ActivityLogUseCase
}

func NewContainer(db *database.DB, secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *Container {
	repos := newRepositories(db)
	return &Container{
		AuthUC:          authusecase.NewAuthUseCase(repos.user, repos.token, secretKey, accessTokenDuration, refreshTokenDuration),
		UserUC:          userusecase.NewUserUseCase(repos.user, repos.role),
		RoleUC:          roleusecase.NewRoleUseCase(repos.role, repos.perm, repos.menu),
		PermUC:          permissionusecase.NewPermissionUseCase(repos.perm),
		MenuUC:          menuusecase.NewMenuUseCase(repos.menu),
		RoomUC:          roomusecase.NewRoomUseCase(repos.room, repos.contract),
		RoomTypeUC:      roomtypeusecase.NewRoomTypeUseCase(repos.roomType),
		TenantUC:        tenantusecase.NewTenantUseCase(repos.tenant, repos.contract),
		ContractUC:      contractusecase.NewContractUseCase(repos.contract, repos.room, repos.tenant),
		DormitoryUC:     dormitoryusecase.NewDormitoryUseCase(repos.dormitory, repos.role),
		ElectricMeterUC: electricmeterusecase.NewElectricMeterUseCase(repos.electricMeter, repos.room),
		WaterMeterUC:    watermeterusecase.NewWaterMeterUseCase(repos.waterMeter, repos.room),
		BillUC:          billusecase.NewBillUseCase(repos.bill, repos.contract),
		DashboardUC:     dashboardusecase.NewDashboardUseCase(repos.dashboard),
		AnalyticsUC:     analyticsusecase.NewAnalyticsUseCase(repos.analytics),
		ExpenseUC:       expenseusecase.NewExpenseUseCase(repos.expense),
		MaintenanceUC:   maintenancerequestusecase.NewMaintenanceRequestUseCase(repos.maintenance, repos.room),
		PaymentUC:       paymentusecase.NewPaymentUseCase(repos.payment, repos.bill),
		AnnouncementUC:  announcementusecase.NewAnnouncementUseCase(repos.announcement),
		ReportUC:        reportusecase.NewReportUseCase(repos.report),
		ParkingUC:       parkingusecase.NewParkingUseCase(repos.parking, repos.tenant),
		ParcelUC:        parcelusecase.NewParcelUseCase(repos.parcel),
		DocumentUC:      documentusecase.NewDocumentUseCase(repos.document, repos.tenant),
		NotificationUC:  notificationusecase.NewNotificationUseCase(repos.notification),
		SearchUC:        searchusecase.NewSearchUseCase(repos.search),
		ActivityLogUC:   alusecase.NewActivityLogUseCase(repos.activityLog),
	}
}
