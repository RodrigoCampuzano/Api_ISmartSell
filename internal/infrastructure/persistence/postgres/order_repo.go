package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/order"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) order.Repository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Save(ctx context.Context, o *order.Order) error {
	q := `INSERT INTO orders
	      (id, buyer_id, business_id, type, status, total, qr_code, delivery_point_id, pickup_deadline)
	      VALUES (:id, :buyer_id, :business_id, :type, :status, :total, :qr_code, :delivery_point_id, :pickup_deadline)`
	_, err := r.db.NamedExecContext(ctx, q, o)
	if err != nil {
		return fmt.Errorf("orderRepo.Save: %w", err)
	}
	return nil
}

func (r *orderRepository) SaveItems(ctx context.Context, items []order.Item) error {
	q := `INSERT INTO order_items (id, order_id, product_id, quantity, unit_price)
	      VALUES (:id, :order_id, :product_id, :quantity, :unit_price)`
	_, err := r.db.NamedExecContext(ctx, q, items)
	return err
}

func (r *orderRepository) Update(ctx context.Context, o *order.Order) error {
	q := `UPDATE orders SET status=:status, qr_code=:qr_code, updated_at=NOW() WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, o)
	return err
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order
	err := r.db.GetContext(ctx, &o, `SELECT * FROM orders WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, order.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("orderRepo.FindByID: %w", err)
	}

	var items []order.Item
	if err := r.db.SelectContext(ctx, &items, `SELECT * FROM order_items WHERE order_id = $1`, id); err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *orderRepository) FindByBuyer(ctx context.Context, buyerID string) ([]*order.Order, error) {
	var list []*order.Order
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM orders WHERE buyer_id = $1 ORDER BY created_at DESC`, buyerID)
	return list, err
}

func (r *orderRepository) FindByBusiness(ctx context.Context, businessID string) ([]*order.Order, error) {
	var list []*order.Order
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM orders WHERE business_id = $1 ORDER BY created_at DESC`, businessID)
	return list, err
}

func (r *orderRepository) FindByQRCode(ctx context.Context, qrCode string) (*order.Order, error) {
	var o order.Order
	err := r.db.GetContext(ctx, &o, `SELECT * FROM orders WHERE qr_code = $1`, qrCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, order.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("orderRepo.FindByQRCode: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) CancelExpired(ctx context.Context) ([]string, error) {
	// Primero obtenemos los IDs de las órdenes que vamos a cancelar
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `
		SELECT id FROM orders
		WHERE type = 'reserved'
		  AND status IN ('reserved','ready')
		  AND pickup_deadline < NOW()`)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.CancelExpired select: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	// Ahora cancelamos todas
	_, err = r.db.ExecContext(ctx, `
		UPDATE orders
		SET status = 'cancelled', updated_at = NOW()
		WHERE type = 'reserved'
		  AND status IN ('reserved','ready')
		  AND pickup_deadline < NOW()`)
	if err != nil {
		return nil, fmt.Errorf("orderRepo.CancelExpired update: %w", err)
	}

	return ids, nil
}
