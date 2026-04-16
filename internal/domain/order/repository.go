package order

import "context"

type Repository interface {
	Save(ctx context.Context, o *Order) error
	Update(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByBuyer(ctx context.Context, buyerID string) ([]*Order, error)
	FindByBusiness(ctx context.Context, businessID string) ([]*Order, error)
	FindByQRCode(ctx context.Context, qrCode string) (*Order, error)
	// CancelExpired cancela pedidos tipo 'reserved' que pasaron su deadline.
	// Retorna los IDs de las órdenes canceladas para procesar reembolsos.
	CancelExpired(ctx context.Context) ([]string, error)
	SaveItems(ctx context.Context, items []Item) error
}
