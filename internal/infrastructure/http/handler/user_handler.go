package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/user"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/response"
)

type UserHandler struct {
	svc services.UserService
}

func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// POST /api/v1/auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	u, token, err := h.svc.Register(r.Context(), services.RegisterInput{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
		Role:     body.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			response.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, user.ErrInvalidRole):
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"user":  u,
		"token": token,
	})
}

// POST /api/v1/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	u, token, err := h.svc.Login(r.Context(), services.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user":  u,
		"token": token,
	})
}

// GET /api/v1/users/me
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromCtx(r.Context())
	u, err := h.svc.GetByID(r.Context(), uid)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	u.Password = "" // nunca exponer el hash
	response.JSON(w, http.StatusOK, u)
}
