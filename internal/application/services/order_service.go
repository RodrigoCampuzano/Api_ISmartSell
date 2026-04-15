package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/business"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/order"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/product"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/qr"
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
	MarkReady(ctx context.Context, id, sellerID string) (*order.Order, error)
	// ScanQR lo llama el vendedor para confirmar entrega.
	ScanQR(ctx context.Context, qrCode, sellerID string) (*order.Order, error)
	CancelOrder(ctx context.Context, id, userID string) (*order.Order, error)
}

// ----- Implementación -----

type orderService struct {
	orderRepo    order.Repository
	productRepo  product.Repository
	businessRepo business.Repository
	notifSvc     NotificationService
	qrSvc        *qr.Service
	paymentSvc   PaymentService
}

func NewOrderService(
	or order.Repository,
	pr product.Repository,
	br business.Repository,
	ns NotificationService,
	qrSvc *qr.Service,
	ps PaymentService,
) OrderService {
	return &orderService{
		orderRepo:    or,
		productRepo:  pr,
		businessRepo: br,
		notifSvc:     ns,
		qrSvc:        qrSvc,
		paymentSvc:   ps,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*order.Order, error) {
	// Validar que el negocio existe
	b, err := s.businessRepo.FindByID(ctx, in.BusinessID)
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
		// Se fuerza a 30 minutos para no alterar el contrato DTO de ReservationHours
		t := time.Now().Add(30 * time.Minute)
		deadline = &t
	}

	o := order.New(uuid.NewString(), in.BuyerID, in.BusinessID,
		order.Type(in.Type), items, in.DeliveryPointID, deadline)
	
	// Generar el código QR de manera incondicional, tal como lo requiere la app.
	o.QRCode = uuid.NewString()

	// Disminuir stock de forma atómica
	for _, it := range items {
		if err := s.productRepo.DecreaseStock(ctx, it.ProductID, it.Quantity); err != nil {
			return nil, err
		}
	}

	// Para compras en línea: el estatus arranca como pagado
	if o.Type == order.TypeOnline {
		_ = o.MarkPaid(o.QRCode)
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

	// Create Mercado Pago Preference
	pref, err := s.paymentSvc.CreatePreference(ctx, o, b.OwnerID)
	if err == nil && pref != nil {
		o.InitPoint = pref.InitPoint
	}

	// Trigger push notification to the seller
	s.notifSvc.SendPushNotification(b.OwnerID, "Nuevo pedido", "Tienes un pedido nuevo en tu tienda")

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

func (s *orderService) MarkReady(ctx context.Context, id, sellerID string) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(ctx, id)
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
	if err := o.MarkReady(); err != nil {
		return nil, err
	}
	if err := s.orderRepo.Update(ctx, o); err != nil {
		return nil, fmt.Errorf("orderService.MarkReady update: %w", err)
	}

	// Trigger push notification to the buyer
	s.notifSvc.SendPushNotification(o.BuyerID, "¡Tu pedido está listo!", "Pasa a recoger tu pedido")

	return o, nil
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

	// Capture the authorized payment in Mercado Pago
	_ = s.paymentSvc.CapturePayment(ctx, o.ID)

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

	// Refund the payment in MP
	_ = s.paymentSvc.CancelPayment(ctx, o.ID)

	return o, nil
}
