package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/business"
)

type businessRepository struct {
	db *sqlx.DB
}

func NewBusinessRepository(db *sqlx.DB) business.Repository {
	return &businessRepository{db: db}
}

func (r *businessRepository) Save(ctx context.Context, b *business.Business) error {
	q := `INSERT INTO businesses (id, owner_id, name, description, type, location, active)
	      VALUES ($1, $2, $3, $4, $5, ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography, $8)`
	_, err := r.db.ExecContext(ctx, q,
		b.ID, b.OwnerID, b.Name, b.Description, b.Type, b.Longitude, b.Latitude, b.Active)
	return err
}

func (r *businessRepository) Update(ctx context.Context, b *business.Business) error {
	q := `UPDATE businesses SET name=$1, description=$2, type=$3,
	      location=ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, active=$6
	      WHERE id=$7`
	_, err := r.db.ExecContext(ctx, q,
		b.Name, b.Description, b.Type, b.Longitude, b.Latitude, b.Active, b.ID)
	return err
}

func (r *businessRepository) FindByID(ctx context.Context, id string) (*business.Business, error) {
	var b business.Business
	q := `SELECT id, owner_id, name, description, type,
	          ST_Y(location::geometry) AS latitude,
	          ST_X(location::geometry) AS longitude,
	          active, created_at, updated_at
	      FROM businesses WHERE id = $1 AND active = TRUE`
	err := r.db.GetContext(ctx, &b, q, id)
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
	q := `SELECT id, owner_id, name, description, type,
	          ST_Y(location::geometry) AS latitude,
	          ST_X(location::geometry) AS longitude,
	          active, created_at, updated_at
	      FROM businesses WHERE owner_id = $1 AND active = TRUE ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &list, q, ownerID)
	return list, err
}

// FindNearby usa PostGIS ST_DWithin para encontrar negocios cercanos.
// radiusKm se convierte a metros porque geography trabaja en metros.
func (r *businessRepository) FindNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*business.Business, error) {
	q := `SELECT id, owner_id, name, description, type,
	          ST_Y(location::geometry) AS latitude,
	          ST_X(location::geometry) AS longitude,
	          active, created_at, updated_at,
	          ST_Distance(location, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography) AS distance
	      FROM businesses
	      WHERE active = TRUE
	        AND ST_DWithin(location, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography, $3)
	      ORDER BY distance`

	var list []*business.Business
	err := r.db.SelectContext(ctx, &list, q, lat, lng, radiusKm*1000)
	return list, err
}

func (r *businessRepository) SaveDeliveryPoint(ctx context.Context, dp *business.DeliveryPoint) error {
	q := `INSERT INTO delivery_points (id, business_id, name, location, active)
	      VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, $6)`
	_, err := r.db.ExecContext(ctx, q,
		dp.ID, dp.BusinessID, dp.Name, dp.Longitude, dp.Latitude, dp.Active)
	return err
}

func (r *businessRepository) FindDeliveryPoints(ctx context.Context, businessID string) ([]*business.DeliveryPoint, error) {
	var list []*business.DeliveryPoint
	q := `SELECT id, business_id, name,
	          ST_Y(location::geometry) AS latitude,
	          ST_X(location::geometry) AS longitude,
	          active, created_at
	      FROM delivery_points WHERE business_id = $1 AND active = TRUE`
	err := r.db.SelectContext(ctx, &list, q, businessID)
	return list, err
}
