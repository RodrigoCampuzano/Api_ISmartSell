package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/product"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/response"
)

type ProductHandler struct {
	svc services.ProductService
}

func NewProductHandler(svc services.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// POST /api/v1/businesses/:businessId/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		ImageURL    string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	ownerID := middleware.UserIDFromCtx(r.Context())
	businessID := chi.URLParam(r, "businessId")

	p, err := h.svc.Create(r.Context(), services.CreateProductInput{
		BusinessID:  businessID,
		OwnerID:     ownerID,
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
		ImageURL:    body.ImageURL,
	})
	if errors.Is(err, product.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, p)
}

// GET /api/v1/businesses/:businessId/products
func (h *ProductHandler) ListByBusiness(w http.ResponseWriter, r *http.Request) {
	businessID := chi.URLParam(r, "businessId")
	list, err := h.svc.GetByBusiness(r.Context(), businessID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/products/:id
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, product.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// PUT /api/v1/products/:id
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		ImageURL    string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	ownerID := middleware.UserIDFromCtx(r.Context())
	p, err := h.svc.Update(r.Context(), services.UpdateProductInput{
		ID:          chi.URLParam(r, "id"),
		OwnerID:     ownerID,
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Stock:       body.Stock,
		ImageURL:    body.ImageURL,
	})
	if errors.Is(err, product.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ownerID := middleware.UserIDFromCtx(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), ownerID)
	if errors.Is(err, product.ErrUnauthorized) {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
