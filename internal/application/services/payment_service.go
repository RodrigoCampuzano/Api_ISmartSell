package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/order"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/domain/payment"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/config"
)

type PreferenceResult struct {
	InitPoint        string `json:"init_point"`
	SandboxInitPoint string `json:"sandbox_init_point"`
}

type PaymentService interface {
	GetAuthorizationURL(sellerID string) string
	HandleOAuthCallback(ctx context.Context, code, sellerID string) error
	CreatePreference(ctx context.Context, o *order.Order, sellerID string) (*PreferenceResult, error)
	CapturePayment(ctx context.Context, orderID string) error
	CancelPayment(ctx context.Context, orderID string) error
	HandleWebhook(ctx context.Context, body []byte) error
}

type paymentService struct {
	cfg            config.Config
	paymentRepo    payment.PaymentRepository
	sellerCredRepo payment.SellerCredentialRepository
	httpClient     *http.Client
}

func NewPaymentService(cfg config.Config, pr payment.PaymentRepository, scr payment.SellerCredentialRepository) PaymentService {
	return &paymentService{
		cfg:            cfg,
		paymentRepo:    pr,
		sellerCredRepo: scr,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *paymentService) GetAuthorizationURL(sellerID string) string {
	return fmt.Sprintf("https://auth.mercadopago.com.mx/authorization?client_id=%s&response_type=code&platform_id=mp&state=id=%s&redirect_uri=%s", s.cfg.MPClientID, sellerID, s.cfg.MPRedirectURI)
}

func (s *paymentService) HandleOAuthCallback(ctx context.Context, code, sellerID string) error {
	payload := map[string]string{
		"client_secret": s.cfg.MPClientSecret,
		"client_id":     s.cfg.MPClientID,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  s.cfg.MPRedirectURI,
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mercadopago.com/oauth/token", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("oauth failed: %s", resp.Status)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		UserID       int64  `json:"user_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	cred := &payment.SellerCredential{
		UserID:       sellerID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		MPUserID:     fmt.Sprintf("%d", tokenResp.UserID),
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return s.sellerCredRepo.Save(ctx, cred)
}

func (s *paymentService) CreatePreference(ctx context.Context, o *order.Order, sellerID string) (*PreferenceResult, error) {
	cred, err := s.sellerCredRepo.FindByUserID(ctx, sellerID)
	if err != nil {
		return nil, fmt.Errorf("seller has not linked MP account: %w", err)
	}

	p := &payment.Payment{
		ID:         uuid.NewString(),
		OrderID:    o.ID,
		Amount:     o.Total,
		Commission: o.Commission(),
		Method:     payment.MethodOnline,
		Status:     payment.StatusPending,
	}

	capture := true
	if o.Type == order.TypeReserved {
		capture = false
		p.Status = payment.StatusAuthorized
	}

	var mpItems []map[string]interface{}
	for _, it := range o.Items {
		priceWithFee := it.UnitPrice * (1 + order.BuyerFeeRate)
		mpItems = append(mpItems, map[string]interface{}{
			"id":         it.ProductID,
			"title":      "Producto",
			"quantity":   it.Quantity,
			"unit_price": math.Round(priceWithFee*100) / 100, // redondear a 2 decimales
		})
	}

	payload := map[string]interface{}{
		"items":              mpItems,
		"marketplace_fee":    o.Commission(),
		"capture":            capture,
		"external_reference": o.ID,
		"payment_methods": map[string]interface{}{
			"excluded_payment_types": []map[string]string{
				{"id": "ticket"},       // OXXO, 7-Eleven, etc.
				{"id": "bank_transfer"},// SPEI
				{"id": "atm"},          // Cajero automático
			},
		},
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mercadopago.com/checkout/preferences", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Use the seller's access token to create the preference on their behalf
	req.Header.Set("Authorization", "Bearer "+cred.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("create preference failed: %d", resp.StatusCode)
	}

	var prefResp struct {
		InitPoint        string `json:"init_point"`
		SandboxInitPoint string `json:"sandbox_init_point"`
		ID               string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prefResp); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Save(ctx, p); err != nil {
		return nil, err
	}

	return &PreferenceResult{
		InitPoint:        prefResp.InitPoint,
		SandboxInitPoint: prefResp.SandboxInitPoint,
	}, nil
}

func (s *paymentService) CapturePayment(ctx context.Context, orderID string) error {
	p, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil || p == nil {
		// Log or return, maybe payment not found locally yet
		return err
	}
	if p.MPPaymentID == "" {
		return fmt.Errorf("no mercado pago payment id linked yet")
	}

	payload := map[string]interface{}{
		"capture": true,
	}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", p.MPPaymentID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.MPAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("capture failed")
	}

	p.Status = payment.StatusCompleted
	return s.paymentRepo.Update(ctx, p)
}

func (s *paymentService) CancelPayment(ctx context.Context, orderID string) error {
	p, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil || p == nil {
		return err
	}

	if p.MPPaymentID == "" {
		return fmt.Errorf("no mercado pago payment id linked yet")
	}

	if p.Status != payment.StatusPending && p.Status != payment.StatusAuthorized {
		return fmt.Errorf("refunds are handled separately, only cancel pre-auths right now")
	}

	// Reembolso parcial: devolver solo el costo de los productos
	// El 2% extra que pagó el comprador se queda como ganancia de la plataforma
	refundAmount := p.Amount // p.Amount = precio productos sin el 2% de recargo
	refundPayload := map[string]interface{}{
		"amount": refundAmount,
	}
	rb, _ := json.Marshal(refundPayload)

	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s/refunds", p.MPPaymentID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(rb))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.MPAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("partial refund failed: %d", resp.StatusCode)
	}

	p.Status = payment.StatusRefunded
	return s.paymentRepo.Update(ctx, p)
}

func (s *paymentService) HandleWebhook(ctx context.Context, body []byte) error {
	// A basic handling of payment notification
	var evt struct {
		Action string `json:"action"`
		Type   string `json:"type"`
		Data   struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}

	if evt.Action == "payment.created" || evt.Type == "payment" {
		url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", evt.Data.ID)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+s.cfg.MPAccessToken)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var paymentResp struct {
			ID                int64  `json:"id"`
			ExternalReference string `json:"external_reference"`
			Status            string `json:"status"`
			StatusDetail      string `json:"status_detail"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err == nil {
			// Find our local payment
			if paymentResp.ExternalReference != "" {
				p, err := s.paymentRepo.FindByOrderID(ctx, paymentResp.ExternalReference)
				if err == nil && p != nil {
					p.MPPaymentID = fmt.Sprintf("%d", paymentResp.ID)
					_ = s.paymentRepo.Update(ctx, p)
				}
			}
		}
	}

	return nil
}
