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
	// ErrPaymentVoided means the payment's receipt was already cancelled.
	ErrPaymentVoided = errors.New("this payment has been voided")
	// ErrReceiptIssued means an edit would change what the issued receipt
	// says (date, methods or amounts); void it and record a new payment.
	ErrReceiptIssued = errors.New("a receipt has been issued for this payment; void it and record a new one to change the date or amounts")
)
