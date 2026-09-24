package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentDormitory is the issuing dormitory as printed on an invoice or
// receipt.
type DocumentDormitory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	PromptPayID string    `json:"promptpay_id"`
}

// DocumentPaymentItem is one method line of a payment (e.g. cash 3,000).
type DocumentPaymentItem struct {
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	ReferenceNo   string  `json:"reference_no"`
}

// DocumentPayment is a payment recorded against the invoice.
type DocumentPayment struct {
	ID          uuid.UUID             `json:"id"`
	PaymentDate time.Time             `json:"payment_date"`
	TotalAmount float64               `json:"total_amount"`
	Note        string                `json:"note"`
	Items       []DocumentPaymentItem `json:"items"`
}

// Document is everything the printable invoice/receipt needs: the invoice
// with its items, who issued it, what has been paid, and, while money is
// still owed and the dormitory has a PromptPay account, the QR payload for
// the outstanding amount.
type Document struct {
	Invoice     Invoice           `json:"invoice"`
	Dormitory   DocumentDormitory `json:"dormitory"`
	Payments    []DocumentPayment `json:"payments"`
	PaidAmount  float64           `json:"paid_amount"`
	Outstanding float64           `json:"outstanding"`
	// PromptPayPayload is the EMVCo string to render as a QR code; empty
	// when nothing is owed, the invoice is cancelled, or no PromptPay
	// account is set on the dormitory.
	PromptPayPayload string `json:"promptpay_payload,omitempty"`
}
