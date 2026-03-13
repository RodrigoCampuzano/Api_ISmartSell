package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/user"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/jwt"
)

// ----- DTOs de entrada -----

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Role     string
}

type LoginInput struct {
	Email    string
	Password string
}

// ----- Puerto de entrada (input port) -----

type UserService interface {
	Register(ctx context.Context, in RegisterInput) (*user.User, string, error)
	Login(ctx context.Context, in LoginInput) (*user.User, string, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
}

// ----- Implementación -----

type userService struct {
	repo   user.Repository
	jwtSvc *jwt.Service
}

func NewUserService(repo user.Repository, jwtSvc *jwt.Service) UserService {
	return &userService{repo: repo, jwtSvc: jwtSvc}
}

func (s *userService) Register(ctx context.Context, in RegisterInput) (*user.User, string, error) {
	exists, err := s.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, "", fmt.Errorf("userService.Register: %w", err)
	}
	if exists {
		return nil, "", user.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("userService.Register hash: %w", err)
	}

	u, err := user.New(uuid.NewString(), in.Name, in.Email, string(hash), user.Role(in.Role))
	if err != nil {
		return nil, "", err
	}

	if err := s.repo.Save(ctx, u); err != nil {
		return nil, "", fmt.Errorf("userService.Register save: %w", err)
	}

	token, err := s.jwtSvc.Generate(u.ID, string(u.Role))
	if err != nil {
		return nil, "", fmt.Errorf("userService.Register jwt: %w", err)
	}

	return u, token, nil
}

func (s *userService) Login(ctx context.Context, in LoginInput) (*user.User, string, error) {
	u, err := s.repo.FindByEmail(ctx, in.Email)
	if err != nil {
		return nil, "", user.ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return nil, "", user.ErrInvalidCreds
	}

	token, err := s.jwtSvc.Generate(u.ID, string(u.Role))
	if err != nil {
		return nil, "", fmt.Errorf("userService.Login jwt: %w", err)
	}

	return u, token, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*user.User, error) {
	return s.repo.FindByID(ctx, id)
}
