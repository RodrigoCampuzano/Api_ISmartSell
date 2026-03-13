package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/product"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) product.Repository {
	return &productRepository{db: db}
}

func (r *productRepository) Save(ctx context.Context, p *product.Product) error {
	q := `INSERT INTO products (id, business_id, name, description, price, stock, image_url, active)
	      VALUES (:id, :business_id, :name, :description, :price, :stock, :image_url, :active)`
	_, err := r.db.NamedExecContext(ctx, q, p)
	return err
}

func (r *productRepository) Update(ctx context.Context, p *product.Product) error {
	q := `UPDATE products SET name=:name, description=:description, price=:price,
	      stock=:stock, image_url=:image_url, active=:active WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, p)
	return err
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET active = FALSE WHERE id = ?`, id)
	return err
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	var p product.Product
	err := r.db.GetContext(ctx, &p, `SELECT * FROM products WHERE id = ? AND active = TRUE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, product.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("productRepo.FindByID: %w", err)
	}
	return &p, nil
}

func (r *productRepository) FindByBusiness(ctx context.Context, businessID string) ([]*product.Product, error) {
	var list []*product.Product
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM products WHERE business_id = ? AND active = TRUE ORDER BY name`, businessID)
	return list, err
}

// DecreaseStock decrementa el stock de forma atómica con UPDATE condicional.
func (r *productRepository) DecreaseStock(ctx context.Context, productID string, qty int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`, qty, productID, qty)
	if err != nil {
		return fmt.Errorf("productRepo.DecreaseStock: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return product.ErrNoStock
	}
	return nil
}
