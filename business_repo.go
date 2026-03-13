package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/pos-app/api/internal/domain/business"
)

type businessRepository struct {
	db *sqlx.DB
}

func NewBusinessRepository(db *sqlx.DB) business.Repository {
	return &businessRepository{db: db}
}

func (r *businessRepository) Save(ctx context.Context, b *business.Business) error {
	q := `INSERT INTO businesses (id, owner_id, name, description, type, latitude, longitude, active)
	      VALUES (:id, :owner_id, :name, :description, :type, :latitude, :longitude, :active)`
	_, err := r.db.NamedExecContext(ctx, q, b)
	return err
}

func (r *businessRepository) Update(ctx context.Context, b *business.Business) error {
	q := `UPDATE businesses SET name=:name, description=:description, type=:type,
	      latitude=:latitude, longitude=:longitude, active=:active WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, b)
	return err
}

func (r *businessRepository) FindByID(ctx context.Context, id string) (*business.Business, error) {
	var b business.Business
	err := r.db.GetContext(ctx, &b, `SELECT * FROM businesses WHERE id = ? AND active = TRUE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, business.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("businessRepo.FindByID: %w", err)
	}
	return &b, nil
}

func (r *businessRepository) FindByOwner(ctx context.Context, ownerID string) ([]*business.Business, error) {
	var list []*business.Business
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM businesses WHERE owner_id = ? AND active = TRUE ORDER BY created_at DESC`, ownerID)
	return list, err
}

// FindNearby usa la fórmula de Haversine aproximada en SQL para encontrar negocios cercanos.
func (r *businessRepository) FindNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*business.Business, error) {
	q := `
	SELECT *, (
	    6371 * ACOS(
	        COS(RADIANS(?)) * COS(RADIANS(latitude)) *
	        COS(RADIANS(longitude) - RADIANS(?)) +
	        SIN(RADIANS(?)) * SIN(RADIANS(latitude))
	    )
	) AS distance
	FROM businesses
	WHERE active = TRUE
	HAVING distance <= ?
	ORDER BY distance`

	var list []*business.Business
	err := r.db.SelectContext(ctx, &list, q, lat, lng, lat, radiusKm)
	return list, err
}

func (r *businessRepository) SaveDeliveryPoint(ctx context.Context, dp *business.DeliveryPoint) error {
	q := `INSERT INTO delivery_points (id, business_id, name, latitude, longitude, active)
	      VALUES (:id, :business_id, :name, :latitude, :longitude, :active)`
	_, err := r.db.NamedExecContext(ctx, q, dp)
	return err
}

func (r *businessRepository) FindDeliveryPoints(ctx context.Context, businessID string) ([]*business.DeliveryPoint, error) {
	var list []*business.DeliveryPoint
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM delivery_points WHERE business_id = ? AND active = TRUE`, businessID)
	return list, err
}
