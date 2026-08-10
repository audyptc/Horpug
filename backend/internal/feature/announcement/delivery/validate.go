package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/announcement/domain"
)

func validateCreateAnnouncementRequest(req *domain.CreateAnnouncementRequest) error {
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

func validateUpdateAnnouncementRequest(req *domain.UpdateAnnouncementRequest) error {
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
