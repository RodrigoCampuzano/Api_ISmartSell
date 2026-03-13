package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/pos-app/api/internal/domain/user"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) user.Repository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	q := `INSERT INTO users (id, name, email, password, role, active)
	      VALUES (:id, :name, :email, :password, :role, :active)`
	_, err := r.db.NamedExecContext(ctx, q, u)
	if err != nil {
		return fmt.Errorf("userRepo.Save: %w", err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = ? AND active = TRUE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByID: %w", err)
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = ? AND active = TRUE`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("userRepo.FindByEmail: %w", err)
	}
	return &u, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users WHERE email = ?`, email)
	return count > 0, err
}
