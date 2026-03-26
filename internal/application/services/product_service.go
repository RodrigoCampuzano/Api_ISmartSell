package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/business"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/product"
)

// ----- DTOs -----

type CreateProductInput struct {
	BusinessID  string
	OwnerID     string
	Name        string
	Description string
	Price       float64
	Stock       int
	ImageURL    string
}

type UpdateProductInput struct {
	ID          string
	OwnerID     string // para verificar ownership vía business
	Name        string
	Description string
	Price       float64
	Stock       int
	ImageURL    string
}

// ----- Puerto de entrada -----

type ProductService interface {
	Create(ctx context.Context, in CreateProductInput) (*product.Product, error)
	Update(ctx context.Context, in UpdateProductInput) (*product.Product, error)
	UpdateStock(ctx context.Context, id, ownerID string, stock int) (*product.Product, error)
	Delete(ctx context.Context, id, ownerID string) error
	GetByID(ctx context.Context, id string) (*product.Product, error)
	GetByBusiness(ctx context.Context, businessID string) ([]*product.Product, error)
}

// ----- Implementación -----

type productService struct {
	productRepo  product.Repository
	businessRepo business.Repository
}

func NewProductService(pr product.Repository, br business.Repository) ProductService {
	return &productService{productRepo: pr, businessRepo: br}
}

func (s *productService) Create(ctx context.Context, in CreateProductInput) (*product.Product, error) {
	b, err := s.businessRepo.FindByID(ctx, in.BusinessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(in.OwnerID) {
		return nil, product.ErrUnauthorized
	}

	p := product.New(uuid.NewString(), in.BusinessID, in.Name, in.Description, in.Price, in.Stock, in.ImageURL)
	if err := s.productRepo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("productService.Create: %w", err)
	}
	return p, nil
}

func (s *productService) Update(ctx context.Context, in UpdateProductInput) (*product.Product, error) {
	p, err := s.productRepo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	b, err := s.businessRepo.FindByID(ctx, p.BusinessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(in.OwnerID) {
		return nil, product.ErrUnauthorized
	}

	p.Name = in.Name
	p.Description = in.Description
	p.Price = in.Price
	p.Stock = in.Stock
	p.ImageURL = in.ImageURL

	if err := s.productRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("productService.Update: %w", err)
	}
	return p, nil
}

func (s *productService) UpdateStock(ctx context.Context, id, ownerID string, stock int) (*product.Product, error) {
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	b, err := s.businessRepo.FindByID(ctx, p.BusinessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(ownerID) {
		return nil, product.ErrUnauthorized
	}

	p.Stock = stock
	if err := s.productRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("productService.UpdateStock: %w", err)
	}
	return p, nil
}

func (s *productService) Delete(ctx context.Context, id, ownerID string) error {
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	b, err := s.businessRepo.FindByID(ctx, p.BusinessID)
	if err != nil {
		return err
	}
	if !b.OwnedBy(ownerID) {
		return product.ErrUnauthorized
	}
	return s.productRepo.Delete(ctx, id)
}

func (s *productService) GetByID(ctx context.Context, id string) (*product.Product, error) {
	return s.productRepo.FindByID(ctx, id)
}

func (s *productService) GetByBusiness(ctx context.Context, businessID string) ([]*product.Product, error) {
	return s.productRepo.FindByBusiness(ctx, businessID)
}
