package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/business"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/response"
)

type BusinessHandler struct {
	svc services.BusinessService
}

func NewBusinessHandler(svc services.BusinessService) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// POST /api/v1/businesses
func (h *BusinessHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Type        string  `json:"type"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	ownerID := middleware.UserIDFromCtx(r.Context())

	b, err := h.svc.Create(r.Context(), services.CreateBusinessInput{
		OwnerID:     ownerID,
		Name:        body.Name,
		Description: body.Description,
		Type:        body.Type,
		Latitude:    body.Latitude,
		Longitude:   body.Longitude,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, b)
}

// GET /api/v1/businesses?lat=&lng=&radius=
func (h *BusinessHandler) ListNearby(w http.ResponseWriter, r *http.Request) {
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radius, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)

	list, err := h.svc.GetNearby(r.Context(), services.NearbyInput{
		Lat: lat, Lng: lng, RadiusKm: radius,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/businesses/mine  (solo vendedor)
func (h *BusinessHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromCtx(r.Context())
	list, err := h.svc.GetByOwner(r.Context(), ownerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/businesses/:id
func (h *BusinessHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	b, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, business.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "business not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, b)
}

// POST /api/v1/businesses/:id/delivery-points
func (h *BusinessHandler) AddDeliveryPoint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	ownerID := middleware.UserIDFromCtx(r.Context())
	businessID := chi.URLParam(r, "id")

	dp, err := h.svc.AddDeliveryPoint(r.Context(), services.CreateDeliveryPointInput{
		BusinessID: businessID,
		OwnerID:    ownerID,
		Name:       body.Name,
		Latitude:   body.Latitude,
		Longitude:  body.Longitude,
	})
	if errors.Is(err, business.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, dp)
}

// DELETE /api/v1/businesses/:id
func (h *BusinessHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ownerID := middleware.UserIDFromCtx(r.Context())

	err := h.svc.Delete(r.Context(), id, ownerID)
	if errors.Is(err, business.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "business not found")
		return
	}
	if errors.Is(err, business.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
