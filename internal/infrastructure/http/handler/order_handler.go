package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/order"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/response"
)

type OrderHandler struct {
	svc services.OrderService
}

func NewOrderHandler(svc services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// POST /api/v1/orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BusinessID      string `json:"business_id"`
		Type            string `json:"type"` // "online" | "reserved"
		DeliveryPointID *string `json:"delivery_point_id,omitempty"`
		ReservationMinutes int  `json:"reservation_minutes,omitempty"`
		Items []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	items := make([]services.OrderItemInput, len(body.Items))
	for i, it := range body.Items {
		items[i] = services.OrderItemInput{ProductID: it.ProductID, Quantity: it.Quantity}
	}

	buyerID := middleware.UserIDFromCtx(r.Context())
	o, err := h.svc.CreateOrder(r.Context(), services.CreateOrderInput{
		BuyerID:          buyerID,
		BusinessID:       body.BusinessID,
		Type:             body.Type,
		Items:            items,
		DeliveryPointID:    body.DeliveryPointID,
		ReservationMinutes: body.ReservationMinutes,
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, o)
}

// GET /api/v1/orders/:id
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	o, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, order.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "order not found")
		return
	}
	if errors.Is(err, order.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, o)
}

// GET /api/v1/orders/my
func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.UserIDFromCtx(r.Context())
	list, err := h.svc.GetByBuyer(r.Context(), buyerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/businesses/:businessId/orders
func (h *OrderHandler) ListByBusiness(w http.ResponseWriter, r *http.Request) {
	sellerID := middleware.UserIDFromCtx(r.Context())
	businessID := chi.URLParam(r, "businessId")
	list, err := h.svc.GetByBusiness(r.Context(), businessID, sellerID)
	if errors.Is(err, order.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// POST /api/v1/orders/:id/ready
func (h *OrderHandler) Ready(w http.ResponseWriter, r *http.Request) {
	sellerID := middleware.UserIDFromCtx(r.Context())
	o, err := h.svc.MarkReady(r.Context(), chi.URLParam(r, "id"), sellerID)
	switch {
	case errors.Is(err, order.ErrNotFound):
		response.Error(w, http.StatusNotFound, "order not found")
	case errors.Is(err, order.ErrUnauthorized):
		response.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, order.ErrInvalidStatus):
		response.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		response.Error(w, http.StatusInternalServerError, err.Error())
	default:
		response.JSON(w, http.StatusOK, o)
	}
}

// POST /api/v1/orders/scan  — vendedor escanea el QR
func (h *OrderHandler) ScanQR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRCode string `json:"qr_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	sellerID := middleware.UserIDFromCtx(r.Context())
	o, err := h.svc.ScanQR(r.Context(), body.QRCode, sellerID)
	switch {
	case errors.Is(err, order.ErrNotFound):
		response.Error(w, http.StatusNotFound, "order not found")
	case errors.Is(err, order.ErrUnauthorized):
		response.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, order.ErrExpired):
		response.Error(w, http.StatusGone, "order expired and cancelled")
	case errors.Is(err, order.ErrInvalidStatus):
		response.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		response.Error(w, http.StatusInternalServerError, err.Error())
	default:
		response.JSON(w, http.StatusOK, o)
	}
}

// POST /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	o, err := h.svc.CancelOrder(r.Context(), chi.URLParam(r, "id"), userID)
	switch {
	case errors.Is(err, order.ErrNotFound):
		response.Error(w, http.StatusNotFound, "order not found")
	case errors.Is(err, order.ErrUnauthorized):
		response.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, order.ErrInvalidStatus):
		response.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		response.Error(w, http.StatusInternalServerError, err.Error())
	default:
		response.JSON(w, http.StatusOK, o)
	}
}
