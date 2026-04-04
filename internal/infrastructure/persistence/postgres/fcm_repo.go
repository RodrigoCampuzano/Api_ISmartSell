package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type FCMRepository struct {
	db *sqlx.DB
}

func NewFCMRepository(db *sqlx.DB) *FCMRepository {
	return &FCMRepository{db: db}
}

func (r *FCMRepository) SaveToken(ctx context.Context, userID, token string) error {
	q := `INSERT INTO fcm_tokens (user_id, token, updated_at) 
	      VALUES ($1, $2, NOW()) 
	      ON CONFLICT (user_id) 
	      DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()`
	
	_, err := r.db.ExecContext(ctx, q, userID, token)
	if err != nil {
		return fmt.Errorf("fcmRepo.SaveToken: %w", err)
	}
	return nil
}

func (r *FCMRepository) GetTokenByUserID(ctx context.Context, userID string) (string, error) {
	var token string
	q := `SELECT token FROM fcm_tokens WHERE user_id = $1`
	err := r.db.GetContext(ctx, &token, q, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("token not found for user %s", userID)
	}
	if err != nil {
		return "", fmt.Errorf("fcmRepo.GetTokenByUserID: %w", err)
	}
	return token, nil
}
