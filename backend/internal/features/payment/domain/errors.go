package domain

import "errors"

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrRequiredPaymentData = errors.New("invoice_id, amount and payment_date are required")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInvalidMethod       = errors.New("invalid payment method")
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvoiceCancelled    = errors.New("cannot record a payment against a cancelled invoice")
)
