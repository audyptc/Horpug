package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/usecase"
)

type Container struct {
	AuthUC        *usecase.AuthUseCase
	UserUC        *usecase.UserUseCase
	RoleUC        *usecase.RoleUseCase
	PermUC        *usecase.PermissionUseCase
	MenuUC        *usecase.MenuUseCase
	RoomUC        *usecase.RoomUseCase
	TenantUC      *usecase.TenantUseCase
	ContractUC    *usecase.ContractUseCase
	MeterReadingUC *usecase.MeterReadingUseCase
	BillUC         *usecase.BillUseCase
}

func NewContainer(db *database.DB, secretKey string) *Container {
	repos := newRepositories(db)
	return &Container{
		AuthUC:         usecase.NewAuthUseCase(repos.user, repos.token, secretKey),
		UserUC:         usecase.NewUserUseCase(repos.user, repos.role),
		RoleUC:         usecase.NewRoleUseCase(repos.role, repos.perm, repos.menu),
		PermUC:         usecase.NewPermissionUseCase(repos.perm),
		MenuUC:         usecase.NewMenuUseCase(repos.menu),
		RoomUC:         usecase.NewRoomUseCase(repos.room),
		TenantUC:       usecase.NewTenantUseCase(repos.tenant),
		ContractUC:     usecase.NewContractUseCase(repos.contract, repos.room, repos.tenant),
		MeterReadingUC: usecase.NewMeterReadingUseCase(repos.meterReading, repos.room),
		BillUC:         usecase.NewBillUseCase(repos.bill, repos.contract),
	}
}
