package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(ps services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: ps}
}

func (h *PaymentHandler) AuthorizeSeller(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	url := h.paymentService.GetAuthorizationURL(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"data": url,
	})
}

func (h *PaymentHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	stateID := r.URL.Query().Get("state") // we sent state=id=sellerID, so state is "id=xxxx"
	
	if len(stateID) > 3 && stateID[:3] == "id=" {
		sellerID := stateID[3:]
		err := h.paymentService.HandleOAuthCallback(r.Context(), code, sellerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Account linked successfully. You may close this window."))
		return
	}

	http.Error(w, "invalid state", http.StatusBadRequest)
}

func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	go func() {
		_ = h.paymentService.HandleWebhook(context.Background(), body)
	}()

	w.WriteHeader(http.StatusOK)
}
