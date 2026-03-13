package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/pos-app/api/internal/domain/business"
	"github.com/pos-app/api/internal/domain/order"
	"github.com/pos-app/api/internal/domain/product"
	"github.com/pos-app/api/pkg/qr"
)

// ----- DTOs -----

type OrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderInput struct {
	BuyerID         string
	BusinessID      string
	Type            string // "online" | "reserved"
	Items           []OrderItemInput
	DeliveryPointID *string
	ReservationHours int // horas límite para recoger (reserved)
}

// ----- Puerto de entrada -----

type OrderService interface {
	CreateOrder(ctx context.Context, in CreateOrderInput) (*order.Order, error)
	GetByID(ctx context.Context, id, userID string) (*order.Order, error)
	GetByBuyer(ctx context.Context, buyerID string) ([]*order.Order, error)
	GetByBusiness(ctx context.Context, businessID, sellerID string) ([]*order.Order, error)
	// ScanQR lo llama el vendedor para confirmar entrega.
	ScanQR(ctx context.Context, qrCode, sellerID string) (*order.Order, error)
	CancelOrder(ctx context.Context, id, userID string) (*order.Order, error)
}

// ----- Implementación -----

type orderService struct {
	orderRepo    order.Repository
	productRepo  product.Repository
	businessRepo business.Repository
	qrSvc        *qr.Service
}

func NewOrderService(
	or order.Repository,
	pr product.Repository,
	br business.Repository,
	qrSvc *qr.Service,
) OrderService {
	return &orderService{
		orderRepo:    or,
		productRepo:  pr,
		businessRepo: br,
		qrSvc:        qrSvc,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*order.Order, error) {
	// Validar que el negocio existe
	_, err := s.businessRepo.FindByID(ctx, in.BusinessID)
	if err != nil {
		return nil, err
	}

	// Construir items y verificar stock
	items := make([]order.Item, 0, len(in.Items))
	for _, it := range in.Items {
		p, err := s.productRepo.FindByID(ctx, it.ProductID)
		if err != nil {
			return nil, err
		}
		if !p.HasStock(it.Quantity) {
			return nil, fmt.Errorf("%w: product %s", product.ErrNoStock, p.Name)
		}
		items = append(items, order.Item{
			ID:        uuid.NewString(),
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: p.Price,
		})
	}

	// Deadline para apartados
	var deadline *time.Time
	if in.Type == string(order.TypeReserved) {
		h := in.ReservationHours
		if h == 0 {
			h = 24
		}
		t := time.Now().Add(time.Duration(h) * time.Hour)
		deadline = &t
	}

	o := order.New(uuid.NewString(), in.BuyerID, in.BusinessID,
		order.Type(in.Type), items, in.DeliveryPointID, deadline)

	// Disminuir stock de forma atómica
	for _, it := range items {
		if err := s.productRepo.DecreaseStock(ctx, it.ProductID, it.Quantity); err != nil {
			return nil, err
		}
	}

	// Para compras en línea: generar QR inmediatamente
	if o.Type == order.TypeOnline {
		code, err := s.qrSvc.Generate(o.ID)
		if err != nil {
			return nil, fmt.Errorf("orderService.CreateOrder qr: %w", err)
		}
		_ = o.MarkPaid(code)
	}

	if err := s.orderRepo.Save(ctx, o); err != nil {
		return nil, fmt.Errorf("orderService.CreateOrder save: %w", err)
	}
	for i := range items {
		items[i].OrderID = o.ID
	}
	if err := s.orderRepo.SaveItems(ctx, items); err != nil {
		return nil, fmt.Errorf("orderService.CreateOrder saveItems: %w", err)
	}

	o.Items = items
	return o, nil
}

func (s *orderService) GetByID(ctx context.Context, id, userID string) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Comprobar que sea el comprador o el vendedor del negocio
	b, err := s.businessRepo.FindByID(ctx, o.BusinessID)
	if err != nil {
		return nil, err
	}
	if !o.BelongsTo(userID) && !b.OwnedBy(userID) {
		return nil, order.ErrUnauthorized
	}
	return o, nil
}

func (s *orderService) GetByBuyer(ctx context.Context, buyerID string) ([]*order.Order, error) {
	return s.orderRepo.FindByBuyer(ctx, buyerID)
}

func (s *orderService) GetByBusiness(ctx context.Context, businessID, sellerID string) ([]*order.Order, error) {
	b, err := s.businessRepo.FindByID(ctx, businessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(sellerID) {
		return nil, order.ErrUnauthorized
	}
	return s.orderRepo.FindByBusiness(ctx, businessID)
}

func (s *orderService) ScanQR(ctx context.Context, qrCode, sellerID string) (*order.Order, error) {
	o, err := s.orderRepo.FindByQRCode(ctx, qrCode)
	if err != nil {
		return nil, err
	}

	b, err := s.businessRepo.FindByID(ctx, o.BusinessID)
	if err != nil {
		return nil, err
	}
	if !b.OwnedBy(sellerID) {
		return nil, order.ErrUnauthorized
	}

	if o.IsExpired() {
		_ = o.Cancel()
		_ = s.orderRepo.Update(ctx, o)
		return nil, order.ErrExpired
	}

	if err := o.MarkDelivered(); err != nil {
		return nil, err
	}
	if err := s.orderRepo.Update(ctx, o); err != nil {
		return nil, fmt.Errorf("orderService.ScanQR update: %w", err)
	}
	return o, nil
}

func (s *orderService) CancelOrder(ctx context.Context, id, userID string) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !o.BelongsTo(userID) {
		return nil, order.ErrUnauthorized
	}
	if err := o.Cancel(); err != nil {
		return nil, err
	}
	if err := s.orderRepo.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
