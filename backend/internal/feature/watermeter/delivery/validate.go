package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/watermeter/domain"
)

func validateCreateWaterMeterRequest(req *domain.CreateWaterMeterRequest) error {
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
