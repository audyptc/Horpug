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
	for _, item := range req.OtherItems {
		if item.Amount < 0 {
			return apierror.BadRequest("other item amount must be >= 0")
		}
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

func validatePaymentSplits(splits []domain.PaymentSplitRequest) error {
	validMethods := map[domain.PaymentMethod]bool{
		domain.PaymentMethodCash:     true,
		domain.PaymentMethodTransfer: true,
		domain.PaymentMethodQR:       true,
	}
	if len(splits) == 0 {
		return apierror.BadRequest("splits must have at least one entry")
	}
	if len(splits) > 3 {
		return apierror.BadRequest("splits must have at most 3 entries")
	}
	seen := map[domain.PaymentMethod]bool{}
	for _, s := range splits {
		if !validMethods[s.Method] {
			return apierror.BadRequest("split method must be one of: cash, transfer, qr")
		}
		if s.Amount <= 0 {
			return apierror.BadRequest("each split amount must be greater than 0")
		}
		if seen[s.Method] {
			return apierror.BadRequest("duplicate split method: " + string(s.Method))
		}
		seen[s.Method] = true
	}
	return nil
}

func CreatePaymentRequest(req *domain.CreatePaymentRequest) error {
	if req.BillID == "" {
		return apierror.BadRequest("bill_id is required")
	}
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return validatePaymentSplits(req.Splits)
}

func UpdatePaymentRequest(req *domain.UpdatePaymentRequest) error {
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return validatePaymentSplits(req.Splits)
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

func CreateParkingSlotRequest(req *domain.CreateParkingSlotRequest) error {
	validVehicleTypes := map[domain.VehicleType]bool{
		domain.VehicleTypeCar:        true,
		domain.VehicleTypeMotorcycle: true,
	}
	validStatuses := map[domain.ParkingStatus]bool{
		domain.ParkingStatusAvailable: true,
		domain.ParkingStatusOccupied:  true,
	}
	if req.SlotNumber == "" {
		return apierror.BadRequest("slot_number is required")
	}
	if !validVehicleTypes[req.VehicleType] {
		return apierror.BadRequest("vehicle_type must be one of: car, motorcycle")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: available, occupied")
	}
	if req.MonthlyFee < 0 {
		return apierror.BadRequest("monthly_fee must be >= 0")
	}
	return nil
}

func UpdateParkingSlotRequest(req *domain.UpdateParkingSlotRequest) error {
	validVehicleTypes := map[domain.VehicleType]bool{
		domain.VehicleTypeCar:        true,
		domain.VehicleTypeMotorcycle: true,
	}
	validStatuses := map[domain.ParkingStatus]bool{
		domain.ParkingStatusAvailable: true,
		domain.ParkingStatusOccupied:  true,
	}
	if req.SlotNumber == "" {
		return apierror.BadRequest("slot_number is required")
	}
	if !validVehicleTypes[req.VehicleType] {
		return apierror.BadRequest("vehicle_type must be one of: car, motorcycle")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: available, occupied")
	}
	if req.MonthlyFee < 0 {
		return apierror.BadRequest("monthly_fee must be >= 0")
	}
	return nil
}

func CreateParcelRequest(req *domain.CreateParcelRequest) error {
	validStatuses := map[domain.ParcelStatus]bool{
		domain.ParcelStatusPending:  true,
		domain.ParcelStatusPickedUp: true,
	}
	if req.RecipientName == "" {
		return apierror.BadRequest("recipient_name is required")
	}
	if req.ReceivedDate.IsZero() {
		return apierror.BadRequest("received_date is required")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: pending, picked_up")
	}
	return nil
}

func UpdateParcelRequest(req *domain.UpdateParcelRequest) error {
	validStatuses := map[domain.ParcelStatus]bool{
		domain.ParcelStatusPending:  true,
		domain.ParcelStatusPickedUp: true,
	}
	if req.RecipientName == "" {
		return apierror.BadRequest("recipient_name is required")
	}
	if req.ReceivedDate.IsZero() {
		return apierror.BadRequest("received_date is required")
	}
	if !validStatuses[req.Status] {
		return apierror.BadRequest("status must be one of: pending, picked_up")
	}
	return nil
}

func CreateDocumentRequest(req *domain.CreateDocumentRequest) error {
	validCategories := map[domain.DocumentCategory]bool{
		domain.DocumentCategoryContract:          true,
		domain.DocumentCategoryIDCard:            true,
		domain.DocumentCategoryHouseRegistration: true,
		domain.DocumentCategoryReceipt:           true,
		domain.DocumentCategoryOther:             true,
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: contract, id_card, house_registration, receipt, other")
	}
	return nil
}

func UpdateDocumentRequest(req *domain.UpdateDocumentRequest) error {
	validCategories := map[domain.DocumentCategory]bool{
		domain.DocumentCategoryContract:          true,
		domain.DocumentCategoryIDCard:            true,
		domain.DocumentCategoryHouseRegistration: true,
		domain.DocumentCategoryReceipt:           true,
		domain.DocumentCategoryOther:             true,
	}
	if req.Title == "" {
		return apierror.BadRequest("title is required")
	}
	if !validCategories[req.Category] {
		return apierror.BadRequest("category must be one of: contract, id_card, house_registration, receipt, other")
	}
	return nil
}

func CreateElectricMeterRequest(req *domain.CreateElectricMeterRequest) error {
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.BillingType != domain.ElectricBillingTypeMeter && req.BillingType != domain.ElectricBillingTypeFlat {
		return apierror.BadRequest("billing_type must be 'meter' or 'flat'")
	}
	if req.ReadingDate.IsZero() {
		return apierror.BadRequest("reading_date is required")
	}
	if req.BillingType == domain.ElectricBillingTypeMeter {
		if req.CurrentReading == nil {
			return apierror.BadRequest("current_reading is required for meter billing")
		}
		if req.UnitPrice == nil || *req.UnitPrice <= 0 {
			return apierror.BadRequest("unit_price must be > 0 for meter billing")
		}
		if *req.CurrentReading < req.PreviousReading {
			return apierror.BadRequest("current_reading must be >= previous_reading")
		}
	} else {
		if req.FlatAmount == nil || *req.FlatAmount <= 0 {
			return apierror.BadRequest("flat_amount must be > 0 for flat billing")
		}
	}
	return nil
}

func CreateWaterMeterRequest(req *domain.CreateWaterMeterRequest) error {
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.BillingType != domain.WaterBillingTypeMeter && req.BillingType != domain.WaterBillingTypeFlat {
		return apierror.BadRequest("billing_type must be 'meter' or 'flat'")
	}
	if req.ReadingDate.IsZero() {
		return apierror.BadRequest("reading_date is required")
	}
	if req.BillingType == domain.WaterBillingTypeMeter {
		if req.CurrentReading == nil {
			return apierror.BadRequest("current_reading is required for meter billing")
		}
		if req.UnitPrice == nil || *req.UnitPrice <= 0 {
			return apierror.BadRequest("unit_price must be > 0 for meter billing")
		}
		prev := 0.0
		if req.PreviousReading != nil {
			prev = *req.PreviousReading
		}
		if *req.CurrentReading < prev {
			return apierror.BadRequest("current_reading must be >= previous_reading")
		}
	} else {
		if req.FlatAmount == nil || *req.FlatAmount <= 0 {
			return apierror.BadRequest("flat_amount must be > 0 for flat billing")
		}
	}
	return nil
}
