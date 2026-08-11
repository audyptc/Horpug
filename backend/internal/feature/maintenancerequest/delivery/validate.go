package delivery

import (
	"apigofiberhorpug/internal/feature/maintenancerequest/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateMaintenanceRequestRequest(req *domain.CreateMaintenanceRequestRequest) error {
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

func validateUpdateMaintenanceRequestRequest(req *domain.UpdateMaintenanceRequestRequest) error {
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
