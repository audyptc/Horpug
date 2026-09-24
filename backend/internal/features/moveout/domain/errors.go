package domain

import "errors"

var (
	ErrContractNotFound  = errors.New("contract not found")
	ErrContractNotActive = errors.New("only an active contract can be moved out")
	ErrAlreadyMovedOut   = errors.New("this contract has already been moved out")
	ErrMoveOutNotFound   = errors.New("move-out not found")
	ErrRequiredDate      = errors.New("move_out_date is required")
	ErrDateBeforeStart   = errors.New("move_out_date must not be before the contract start date")
	ErrInvalidItem       = errors.New("each deduction needs a description and an amount greater than zero")
	ErrDormitoryNotFound = errors.New("dormitory not found")
	ErrInvalidPreset     = errors.New("each preset needs a name and an amount greater than zero")
)
