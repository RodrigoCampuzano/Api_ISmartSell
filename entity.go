package user

import (
	"errors"
	"time"
)

type Role string

const (
	RoleSeller Role = "seller"
	RoleBuyer  Role = "buyer"
)

var (
	ErrNotFound      = errors.New("user: not found")
	ErrEmailTaken    = errors.New("user: email already in use")
	ErrInvalidRole   = errors.New("user: invalid role")
	ErrInvalidCreds  = errors.New("user: invalid credentials")
)

// User es el aggregate root del dominio de usuarios.
type User struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"` // bcrypt hash
	Role      Role      `db:"role"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func New(id, name, email, hashedPassword string, role Role) (*User, error) {
	if role != RoleSeller && role != RoleBuyer {
		return nil, ErrInvalidRole
	}
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		Active:   true,
	}, nil
}

func (u *User) IsSeller() bool { return u.Role == RoleSeller }
func (u *User) IsBuyer() bool  { return u.Role == RoleBuyer }
