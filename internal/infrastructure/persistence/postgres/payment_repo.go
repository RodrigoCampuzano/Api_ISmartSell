package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/payment"
)

type paymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) payment.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	q := `INSERT INTO payments (id, order_id, amount, commission, method, status, mp_payment_id)
	      VALUES (:id, :order_id, :amount, :commission, :method, :status, :mp_payment_id)`
	_, err := r.db.NamedExecContext(ctx, q, p)
	return err
}

func (r *paymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	q := `UPDATE payments SET status=:status, mp_payment_id=:mp_payment_id, updated_at=NOW() WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, p)
	return err
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	var p payment.Payment
	err := r.db.GetContext(ctx, &p, `SELECT * FROM payments WHERE order_id = $1`, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("payment not found")
	}
	return &p, err
}

func (r *paymentRepository) FindByMPPaymentID(ctx context.Context, mpPaymentID string) (*payment.Payment, error) {
	var p payment.Payment
	err := r.db.GetContext(ctx, &p, `SELECT * FROM payments WHERE mp_payment_id = $1`, mpPaymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("payment not found")
	}
	return &p, err
}

type sellerCredentialRepo struct {
	db *sqlx.DB
}

func NewSellerCredentialRepository(db *sqlx.DB) payment.SellerCredentialRepository {
	return &sellerCredentialRepo{db: db}
}

func (r *sellerCredentialRepo) Save(ctx context.Context, cred *payment.SellerCredential) error {
	q := `INSERT INTO seller_mp_credentials (user_id, access_token, refresh_token, mp_user_id, expires_at)
	      VALUES (:user_id, :access_token, :refresh_token, :mp_user_id, :expires_at)
	      ON CONFLICT (user_id) DO UPDATE SET 
	      access_token = EXCLUDED.access_token,
	      refresh_token = EXCLUDED.refresh_token,
	      mp_user_id = EXCLUDED.mp_user_id,
	      expires_at = EXCLUDED.expires_at,
	      updated_at = NOW()`
	_, err := r.db.NamedExecContext(ctx, q, cred)
	return err
}

func (r *sellerCredentialRepo) FindByUserID(ctx context.Context, userID string) (*payment.SellerCredential, error) {
	var cred payment.SellerCredential
	err := r.db.GetContext(ctx, &cred, `SELECT * FROM seller_mp_credentials WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("credentials not found")
	}
	return &cred, err
}
