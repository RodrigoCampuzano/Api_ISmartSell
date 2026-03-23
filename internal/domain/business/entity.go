package business

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("business: not found")
	ErrUnauthorized  = errors.New("business: not owner")
)

type Business struct {
	ID          string         `db:"id"         json:"id"`
	OwnerID     string         `db:"owner_id"   json:"owner_id"`
	Name        string         `db:"name"       json:"name"`
	Description string         `db:"description" json:"description"`
	Type        string         `db:"type"       json:"type"`
	Latitude    float64        `db:"latitude"   json:"latitude"`
	Longitude   float64        `db:"longitude"  json:"longitude"`
	Active      bool           `db:"active"     json:"active"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	Distance    float64        `db:"distance"   json:"distance,omitempty"` // metros, solo en FindNearby

	DeliveryPoints []DeliveryPoint `db:"-" json:"delivery_points,omitempty"`
}

type DeliveryPoint struct {
	ID         string    `db:"id"`
	BusinessID string    `db:"business_id"`
	Name       string    `db:"name"`
	Latitude   float64   `db:"latitude"`
	Longitude  float64   `db:"longitude"`
	Active     bool      `db:"active"`
	CreatedAt  time.Time `db:"created_at"`
}

func New(id, ownerID, name, description, bType string, lat, lng float64) *Business {
	return &Business{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		Type:        bType,
		Latitude:    lat,
		Longitude:   lng,
		Active:      true,
	}
}

func (b *Business) OwnedBy(userID string) bool {
	return b.OwnerID == userID
}
