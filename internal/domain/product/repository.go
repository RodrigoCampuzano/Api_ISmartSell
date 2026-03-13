package product

import "context"

type Repository interface {
	Save(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Product, error)
	FindByBusiness(ctx context.Context, businessID string) ([]*Product, error)
	// DecreaseStock disminuye el stock de forma atómica (usado en creación de órdenes).
	DecreaseStock(ctx context.Context, productID string, qty int) error
}
