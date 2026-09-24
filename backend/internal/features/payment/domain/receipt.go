package domain

import "github.com/google/uuid"

// ReceiptDormitory is the issuing dormitory as printed on a receipt.
type ReceiptDormitory struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Phone   string    `json:"phone"`
}

// ReceiptInvoice is the invoice a receipt pays, with its balance counted over
// the invoice's active (not voided) payments.
type ReceiptInvoice struct {
	ID          uuid.UUID `json:"id"`
	InvoiceNo   string    `json:"invoice_no"`
	PeriodYear  int       `json:"period_year"`
	PeriodMonth int       `json:"period_month"`
	TotalAmount float64   `json:"total_amount"`
	PaidAmount  float64   `json:"paid_amount"`
	Outstanding float64   `json:"outstanding"`
}

// Receipt is everything the printable receipt needs.
type Receipt struct {
	Payment   Payment          `json:"payment"`
	Dormitory ReceiptDormitory `json:"dormitory"`
	Invoice   ReceiptInvoice   `json:"invoice"`
}
