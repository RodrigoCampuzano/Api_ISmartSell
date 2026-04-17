package order

import (
	"errors"
	"time"
)

type Status string
type Type string

const (
	StatusPending    Status = "pending"
	StatusPaid       Status = "paid"
	StatusReserved   Status = "reserved"
	StatusReady      Status = "ready"
	StatusDelivered  Status = "delivered"
	StatusCancelled  Status = "cancelled"
	StatusAuthorized Status = "authorized"

	TypeOnline   Type = "online"
	TypeReserved Type = "reserved"

	CommissionRate    = 0.05 // 5% total para la plataforma
	BuyerFeeRate      = 0.02 // 2% lo absorbe el comprador
	SellerFeeRate     = 0.03 // 3% lo absorbe el vendedor
)

var (
	ErrNotFound      = errors.New("order: not found")
	ErrUnauthorized  = errors.New("order: unauthorized")
	ErrInvalidStatus = errors.New("order: invalid status transition")
	ErrExpired       = errors.New("order: pickup deadline exceeded")
)

type Item struct {
	ID        string  `db:"id" json:"id"`
	OrderID   string  `db:"order_id" json:"order_id"`
	ProductID string  `db:"product_id" json:"product_id"`
	Quantity  int     `db:"quantity" json:"quantity"`
	UnitPrice float64 `db:"unit_price" json:"unit_price"`
}

func (i *Item) Subtotal() float64 { return i.UnitPrice * float64(i.Quantity) }

type Order struct {
	ID              string     `db:"id" json:"id"`
	BuyerID         string     `db:"buyer_id" json:"buyer_id"`
	BusinessID      string     `db:"business_id" json:"business_id"`
	Type            Type       `db:"type" json:"type"`
	Status          Status     `db:"status" json:"status"`
	Total           float64    `db:"total" json:"total"`
	QRCode          string     `db:"qr_code" json:"qr_code"`
	DeliveryPointID *string    `db:"delivery_point_id" json:"delivery_point_id,omitempty"`
	PickupDeadline  *time.Time `db:"pickup_deadline" json:"pickup_deadline,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
	Items           []Item     `db:"-" json:"items"`
	InitPoint       string     `db:"-" json:"init_point,omitempty"`
}

func New(id, buyerID, businessID string, orderType Type, items []Item, dpID *string, deadline *time.Time) *Order {
	total := 0.0
	for _, it := range items {
		total += it.Subtotal()
	}

	status := StatusPending

	return &Order{
		ID:              id,
		BuyerID:         buyerID,
		BusinessID:      businessID,
		Type:            orderType,
		Status:          status,
		Total:           total,
		DeliveryPointID: dpID,
		PickupDeadline:  deadline,
		Items:           items,
	}
}

func (o *Order) Commission() float64 { return o.Total * CommissionRate }

func (o *Order) MarkPaid(qrCode string) error {
	if o.Status != StatusPending {
		return ErrInvalidStatus
	}
	o.Status = StatusPaid
	o.QRCode = qrCode
	return nil
}

func (o *Order) MarkReady() error {
	if o.Status != StatusPaid && o.Status != StatusReserved {
		return ErrInvalidStatus
	}
	o.Status = StatusReady
	return nil
}

func (o *Order) MarkDelivered() error {
	if o.Status != StatusReady {
		return ErrInvalidStatus
	}
	o.Status = StatusDelivered
	return nil
}

func (o *Order) Cancel() error {
	if o.Status == StatusDelivered || o.Status == StatusCancelled {
		return ErrInvalidStatus
	}
	o.Status = StatusCancelled
	return nil
}

func (o *Order) IsExpired() bool {
	if o.PickupDeadline == nil {
		return false
	}
	return time.Now().After(*o.PickupDeadline)
}

func (o *Order) BelongsTo(userID string) bool { return o.BuyerID == userID }
