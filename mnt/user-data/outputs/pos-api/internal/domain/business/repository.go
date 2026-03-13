package business

import "context"

type Repository interface {
	Save(ctx context.Context, b *Business) error
	Update(ctx context.Context, b *Business) error
	FindByID(ctx context.Context, id string) (*Business, error)
	FindByOwner(ctx context.Context, ownerID string) ([]*Business, error)
	// FindNearby devuelve negocios dentro de radiusKm kilómetros.
	FindNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*Business, error)
	SaveDeliveryPoint(ctx context.Context, dp *DeliveryPoint) error
	FindDeliveryPoints(ctx context.Context, businessID string) ([]*DeliveryPoint, error)
}
