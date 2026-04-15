package payment

import (
	"time"
)

type Status string
type Method string

const (
	StatusPending    Status = "pending"
	StatusAuthorized Status = "authorized"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusRefunded   Status = "refunded"
	StatusCancelled  Status = "cancelled"

	MethodOnline Method = "online"
	MethodCash   Method = "cash"
)

// Payment represents a single charge or pre-authorization attempt
type Payment struct {
	ID          string    `db:"id" json:"id"`
	OrderID     string    `db:"order_id" json:"order_id"`
	Amount      float64   `db:"amount" json:"amount"`
	Commission  float64   `db:"commission" json:"commission"`
	Method      Method    `db:"method" json:"method"`
	Status      Status    `db:"status" json:"status"`
	MPPaymentID string    `db:"mp_payment_id" json:"mp_payment_id,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// SellerCredential represents the Mercado Pago OAuth tokens for a seller
type SellerCredential struct {
	UserID       string    `db:"user_id" json:"user_id"`
	AccessToken  string    `db:"access_token" json:"-"`
	RefreshToken string    `db:"refresh_token" json:"-"`
	MPUserID     string    `db:"mp_user_id" json:"mp_user_id"`
	ExpiresAt    time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
