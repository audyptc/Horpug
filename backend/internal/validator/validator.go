package validator

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
)

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

