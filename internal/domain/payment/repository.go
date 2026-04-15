package payment

import "context"

type PaymentRepository interface {
	Save(ctx context.Context, p *Payment) error
	Update(ctx context.Context, p *Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	FindByMPPaymentID(ctx context.Context, mpPaymentID string) (*Payment, error)
}

type SellerCredentialRepository interface {
	Save(ctx context.Context, cred *SellerCredential) error
	FindByUserID(ctx context.Context, userID string) (*SellerCredential, error)
}
