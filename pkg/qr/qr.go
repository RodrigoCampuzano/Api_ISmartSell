package qr

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// Service genera QR codes como PNG en base64.
type Service struct{}

func NewService() *Service { return &Service{} }

// Generate devuelve un string único para el pedido y el PNG en base64.
// El "token" del QR es simplemente el orderID; en producción podrías
// añadir una firma HMAC para mayor seguridad.
func (s *Service) Generate(orderID string) (string, error) {
	png, err := qrcode.Encode(orderID, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("qr.Generate: %w", err)
	}
	// Guardamos el orderID como el "código" (escaneado vía texto).
	_ = base64.StdEncoding.EncodeToString(png) // imagen disponible si la necesitas
	return orderID, nil
}

// GenerateBase64 devuelve el PNG del QR en base64 (para la app móvil).
func (s *Service) GenerateBase64(orderID string) (string, error) {
	png, err := qrcode.Encode(orderID, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("qr.GenerateBase64: %w", err)
	}
	return base64.StdEncoding.EncodeToString(png), nil
}
