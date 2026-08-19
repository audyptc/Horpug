package domain

import "errors"

var (
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrRequiredInvoiceData  = errors.New("contract_id, period_year, period_month, issue_date and due_date are required")
	ErrInvalidInvoicePeriod = errors.New("period_month must be between 1 and 12")
	ErrInvalidInvoiceDates  = errors.New("due_date must not be before issue_date")
	ErrInvalidInvoiceStatus = errors.New("invalid invoice status")
	ErrContractNotFound     = errors.New("contract not found")
	ErrInvoiceExists        = errors.New("an invoice for this contract and period already exists")

	ErrRequiredInvoiceItemData  = errors.New("description and amount are required")
	ErrInvalidInvoiceItemAmount = errors.New("amount must be greater than zero")
	ErrInvoiceItemNotFound      = errors.New("invoice item not found")
	ErrInvoiceItemNotRemovable  = errors.New("only manually added items can be removed")
	ErrInvoiceLocked            = errors.New("invoice items cannot be changed once the invoice is paid or cancelled")

	ErrTenantLineNotLinked = errors.New("tenant has not linked a LINE account")
)
