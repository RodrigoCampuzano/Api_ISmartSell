package user

import "context"

// Repository es el puerto de salida para persistencia de usuarios.
// La implementación vive en infrastructure/persistence/mysql.
type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
