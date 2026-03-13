package product

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("product: not found")
	ErrNoStock      = errors.New("product: insufficient stock")
	ErrUnauthorized = errors.New("product: not owner")
)

type Product struct {
	ID          string    `db:"id"`
	BusinessID  string    `db:"business_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	ImageURL    string    `db:"image_url"`
	Active      bool      `db:"active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func New(id, businessID, name, description string, price float64, stock int, imageURL string) *Product {
	return &Product{
		ID:          id,
		BusinessID:  businessID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		ImageURL:    imageURL,
		Active:      true,
	}
}

func (p *Product) HasStock(qty int) bool {
	return p.Stock >= qty
}

func (p *Product) DecreaseStock(qty int) error {
	if !p.HasStock(qty) {
		return ErrNoStock
	}
	p.Stock -= qty
	return nil
}
