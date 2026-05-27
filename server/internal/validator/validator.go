package validator

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
)

func LoginRequest(req *domain.LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return apierror.BadRequest("email and password are required")
	}
	return nil
}

func RefreshRequest(req *domain.RefreshRequest) error {
	if req.RefreshToken == "" {
		return apierror.BadRequest("refresh token is required")
	}
	return nil
}

func CreateUserRequest(req *domain.CreateUserRequest) error {
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return apierror.BadRequest("full_name, email and password are required")
	}
	return nil
}

func AssignRoleRequest(req *domain.AssignRoleRequest) error {
	if req.RoleID == "" {
		return apierror.BadRequest("role_id is required")
	}
	return nil
}

func CreateRoleRequest(req *domain.CreateRoleRequest) error {
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}
	return nil
}

func CreateTenantRequest(req *domain.CreateTenantRequest) error {
	if req.FirstName == "" || req.LastName == "" {
		return apierror.BadRequest("first_name and last_name are required")
	}
	if req.Phone == "" {
		return apierror.BadRequest("phone is required")
	}
	if req.IDCard == "" {
		return apierror.BadRequest("id_card is required")
	}
	return nil
}

func CreateContractRequest(req *domain.CreateContractRequest) error {
	if req.TenantID == "" {
		return apierror.BadRequest("tenant_id is required")
	}
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.StartDate.IsZero() {
		return apierror.BadRequest("start_date is required")
	}
	if req.RentPrice <= 0 {
		return apierror.BadRequest("rent_price must be greater than 0")
	}
	return nil
}

func CreateRoomRequest(req *domain.CreateRoomRequest) error {
	if req.RoomNumber == "" {
		return apierror.BadRequest("room_number is required")
	}
	if req.Floor <= 0 {
		return apierror.BadRequest("floor must be greater than 0")
	}
	if req.RentPrice <= 0 {
		return apierror.BadRequest("rent_price must be greater than 0")
	}
	return nil
}

func CreateBillRequest(req *domain.CreateBillRequest) error {
	if req.ContractID == "" {
		return apierror.BadRequest("contract_id is required")
	}
	if req.BillingMonth.IsZero() {
		return apierror.BadRequest("billing_month is required")
	}
	if req.RentAmount < 0 {
		return apierror.BadRequest("rent_amount must be >= 0")
	}
	return nil
}

func CreateExpenseRequest(req *domain.CreateExpenseRequest) error {
	validCategories := map[domain.ExpenseCategory]bool{
		domain.ExpenseCategoryRepair:    true,
		domain.ExpenseCategoryUtilities: true,
		domain.ExpenseCategorySupplies:  true,
		domain.ExpenseCategorySalary:    true,
		domain.ExpenseCategoryOther:     true,
	}
	if req.Description == "" {
		return apierror.BadRequest("description is required")
	}
	if req.ExpenseDate.IsZero() {
		return apierror.BadRequest("expense_date is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: repair, utilities, supplies, salary, other")
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	return nil
}

func UpdateExpenseRequest(req *domain.UpdateExpenseRequest) error {
	validCategories := map[domain.ExpenseCategory]bool{
		domain.ExpenseCategoryRepair:    true,
		domain.ExpenseCategoryUtilities: true,
		domain.ExpenseCategorySupplies:  true,
		domain.ExpenseCategorySalary:    true,
		domain.ExpenseCategoryOther:     true,
	}
	if req.Description == "" {
		return apierror.BadRequest("description is required")
	}
	if req.ExpenseDate.IsZero() {
		return apierror.BadRequest("expense_date is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: repair, utilities, supplies, salary, other")
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	return nil
}

func CreateMaintenanceRequestRequest(req *domain.CreateMaintenanceRequestRequest) error {
	validStatuses := map[domain.MaintenanceStatus]bool{
		domain.MaintenanceStatusOpen:       true,
		domain.MaintenanceStatusInProgress: true,
		domain.MaintenanceStatusDone:       true,
		domain.MaintenanceStatusCancelled:  true,
	}
	validPriorities := map[domain.MaintenancePriority]bool{
		domain.MaintenancePriorityLow:    true,
		domain.MaintenancePriorityNormal: true,
		domain.MaintenancePriorityHigh:   true,
		domain.MaintenancePriorityUrgent: true,
	}
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if req.ReportedDate.IsZero() {
		return apierror.BadRequest("reported_date is required")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: open, in_progress, done, cancelled")
	}
	if !validPriorities[req.Priority] {
		return apierror.BadRequest("priority must be one of: low, normal, high, urgent")
	}
	return nil
}

func UpdateMaintenanceRequestRequest(req *domain.UpdateMaintenanceRequestRequest) error {
	validStatuses := map[domain.MaintenanceStatus]bool{
		domain.MaintenanceStatusOpen:       true,
		domain.MaintenanceStatusInProgress: true,
		domain.MaintenanceStatusDone:       true,
		domain.MaintenanceStatusCancelled:  true,
	}
	validPriorities := map[domain.MaintenancePriority]bool{
		domain.MaintenancePriorityLow:    true,
		domain.MaintenancePriorityNormal: true,
		domain.MaintenancePriorityHigh:   true,
		domain.MaintenancePriorityUrgent: true,
	}
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if req.ReportedDate.IsZero() {
		return apierror.BadRequest("reported_date is required")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: open, in_progress, done, cancelled")
	}
	if !validPriorities[req.Priority] {
		return apierror.BadRequest("priority must be one of: low, normal, high, urgent")
	}
	return nil
}

func CreatePaymentRequest(req *domain.CreatePaymentRequest) error {
	validMethods := map[domain.PaymentMethod]bool{
		domain.PaymentMethodCash:     true,
		domain.PaymentMethodTransfer: true,
		domain.PaymentMethodQR:       true,
	}
	if req.BillID == "" {
		return apierror.BadRequest("bill_id is required")
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	if !validMethods[req.Method] {
		return apierror.BadRequest("method must be one of: cash, transfer, qr")
	}
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return nil
}

func UpdatePaymentRequest(req *domain.UpdatePaymentRequest) error {
	validMethods := map[domain.PaymentMethod]bool{
		domain.PaymentMethodCash:     true,
		domain.PaymentMethodTransfer: true,
		domain.PaymentMethodQR:       true,
	}
	if req.Amount <= 0 {
		return apierror.BadRequest("amount must be greater than 0")
	}
	if !validMethods[req.Method] {
		return apierror.BadRequest("method must be one of: cash, transfer, qr")
	}
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return nil
}

func CreateAnnouncementRequest(req *domain.CreateAnnouncementRequest) error {
	validTypes := map[domain.AnnouncementType]bool{
		domain.AnnouncementTypeGeneral:     true,
		domain.AnnouncementTypeMaintenance: true,
		domain.AnnouncementTypePayment:     true,
		domain.AnnouncementTypeEmergency:   true,
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if req.Content == "" {
		return apierror.BadRequest("content is required")
	}
	if !validTypes[req.Type] {
		return apierror.BadRequest("type must be one of: general, maintenance, payment, emergency")
	}
	if req.PublishedAt.IsZero() {
		return apierror.BadRequest("published_at is required")
	}
	return nil
}

func UpdateAnnouncementRequest(req *domain.UpdateAnnouncementRequest) error {
	validTypes := map[domain.AnnouncementType]bool{
		domain.AnnouncementTypeGeneral:     true,
		domain.AnnouncementTypeMaintenance: true,
		domain.AnnouncementTypePayment:     true,
		domain.AnnouncementTypeEmergency:   true,
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if req.Content == "" {
		return apierror.BadRequest("content is required")
	}
	if !validTypes[req.Type] {
		return apierror.BadRequest("type must be one of: general, maintenance, payment, emergency")
	}
	if req.PublishedAt.IsZero() {
		return apierror.BadRequest("published_at is required")
	}
	return nil
}

func CreateMeterReadingRequest(req *domain.CreateMeterReadingRequest) error {
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.MeterType != domain.MeterTypeElectric && req.MeterType != domain.MeterTypeWater {
		return apierror.BadRequest("meter_type must be 'electric' or 'water'")
	}
	if req.ReadingDate.IsZero() {
		return apierror.BadRequest("reading_date is required")
	}
	if req.CurrentReading < req.PreviousReading {
		return apierror.BadRequest("current_reading must be >= previous_reading")
	}
	if req.UnitPrice <= 0 {
		return apierror.BadRequest("unit_price must be greater than 0")
	}
	return nil
}
