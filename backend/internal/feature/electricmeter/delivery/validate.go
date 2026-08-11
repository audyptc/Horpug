package delivery

import (
	"apigofiberhorpug/internal/feature/electricmeter/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateElectricMeterRequest(req *domain.CreateElectricMeterRequest) error {
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
