package domain

import "errors"

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrRequiredPaymentData = errors.New("invoice_id and payment_date are required")
	ErrRequiredItems       = errors.New("at least one payment item is required")
	ErrInvalidAmount       = errors.New("each payment item's amount must be greater than zero")
	ErrInvalidMethod       = errors.New("invalid payment method")
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvoiceCancelled    = errors.New("cannot record a payment against a cancelled invoice")
	// ErrPaymentExceedsInvoice means the invoice's recorded payments would add
	// up to more than its total_amount.
	ErrPaymentExceedsInvoice = errors.New("recorded payments cannot exceed the invoice total")
)
