package delivery

import (
	"apigofiberhorpug/internal/feature/room/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateRoomRequest(req *domain.CreateRoomRequest) error {
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
