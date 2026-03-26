package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/business"
)

// ----- DTOs -----

type CreateBusinessInput struct {
	OwnerID     string
	Name        string
	Description string
	Type        string
	Latitude    float64
	Longitude   float64
}

type CreateDeliveryPointInput struct {
	BusinessID string
	OwnerID    string
	Name       string
	Latitude   float64
	Longitude  float64
}

type NearbyInput struct {
	Lat      float64
	Lng      float64
	RadiusKm float64
}

// ----- Puerto de entrada -----

type BusinessService interface {
	Create(ctx context.Context, in CreateBusinessInput) (*business.Business, error)
	GetByID(ctx context.Context, id string) (*business.Business, error)
	GetByOwner(ctx context.Context, ownerID string) ([]*business.Business, error)
	GetNearby(ctx context.Context, in NearbyInput) ([]*business.Business, error)
	AddDeliveryPoint(ctx context.Context, in CreateDeliveryPointInput) (*business.DeliveryPoint, error)
	Delete(ctx context.Context, id, ownerID string) error
}

// ----- Implementación -----

type businessService struct {
	repo business.Repository
}

func NewBusinessService(repo business.Repository) BusinessService {
	return &businessService{repo: repo}
}

func (s *businessService) Create(ctx context.Context, in CreateBusinessInput) (*business.Business, error) {
	b := business.New(uuid.NewString(), in.OwnerID, in.Name, in.Description, in.Type, in.Latitude, in.Longitude)
	if err := s.repo.Save(ctx, b); err != nil {
		return nil, fmt.Errorf("businessService.Create: %w", err)
	}
	return b, nil
}

func (s *businessService) GetByID(ctx context.Context, id string) (*business.Business, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Cargar puntos de entrega
	dps, err := s.repo.FindDeliveryPoints(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, dp := range dps {
		b.DeliveryPoints = append(b.DeliveryPoints, *dp)
	}
	return b, nil
}

func (s *businessService) GetByOwner(ctx context.Context, ownerID string) ([]*business.Business, error) {
	return s.repo.FindByOwner(ctx, ownerID)
}

func (s *businessService) GetNearby(ctx context.Context, in NearbyInput) ([]*business.Business, error) {
	radius := in.RadiusKm
	if radius == 0 {
		radius = 5
	}
	return s.repo.FindNearby(ctx, in.Lat, in.Lng, radius)
}

func (s *businessService) AddDeliveryPoint(ctx context.Context, in CreateDeliveryPointInput) (*business.DeliveryPoint, error) {
	b, err := s.repo.FindByID(ctx, in.BusinessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(in.OwnerID) {
		return nil, business.ErrUnauthorized
	}

	dp := &business.DeliveryPoint{
		ID:         uuid.NewString(),
		BusinessID: in.BusinessID,
		Name:       in.Name,
		Latitude:   in.Latitude,
		Longitude:  in.Longitude,
		Active:     true,
	}
	if err := s.repo.SaveDeliveryPoint(ctx, dp); err != nil {
		return nil, fmt.Errorf("businessService.AddDeliveryPoint: %w", err)
	}
	return dp, nil
}

func (s *businessService) Delete(ctx context.Context, id, ownerID string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !b.OwnedBy(ownerID) {
		return business.ErrUnauthorized
	}
	b.Active = false
	if err := s.repo.Update(ctx, b); err != nil {
		return fmt.Errorf("businessService.Delete: %w", err)
	}
	return nil
}
