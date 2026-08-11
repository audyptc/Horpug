package delivery

import (
	"apigofiberhorpug/internal/feature/document/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateDocumentRequest(req *domain.CreateDocumentRequest) error {
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

func validateUpdateDocumentRequest(req *domain.UpdateDocumentRequest) error {
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
