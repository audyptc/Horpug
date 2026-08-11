package delivery

import (
	"apigofiberhorpug/internal/features/parcel/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateParcelRequest(req *domain.CreateParcelRequest) error {
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

func validateUpdateParcelRequest(req *domain.UpdateParcelRequest) error {
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
