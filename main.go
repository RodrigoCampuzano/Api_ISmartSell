package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/application/services"
	infraHTTP "github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/handler"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/persistence/postgres"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/config"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/jwt"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/qr"
)

func main() {
	cfg := config.Load()

	// ── Base de datos ───────────────────────────────────────
	db, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	// ── Servicios compartidos ───────────────────────────────
	jwtSvc := jwt.NewService(cfg.JWTSecret, cfg.JWTTTLHrs)
	qrSvc := qr.NewService()

	// ── Repositorios (adaptadores de salida) ────────────────
	userRepo := postgres.NewUserRepository(db)
	businessRepo := postgres.NewBusinessRepository(db)
	productRepo := postgres.NewProductRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	fcmRepo := postgres.NewFCMRepository(db)
	paymentRepo := postgres.NewPaymentRepository(db)
	sellerCredRepo := postgres.NewSellerCredentialRepository(db)

	// ── Servicios de aplicación (casos de uso) ──────────────
	userSvc := services.NewUserService(userRepo, jwtSvc)
	businessSvc := services.NewBusinessService(businessRepo)
	productSvc := services.NewProductService(productRepo, businessRepo)

	notifSvc, err := services.NewNotificationService(fcmRepo, "/var/www/Api_ISmartSell/ismartshell-firebase-adminsdk-fbsvc-db9cfd3ead.json")
	if err != nil {
		log.Fatalf("notification service init: %v", err)
	}

	paymentSvc := services.NewPaymentService(cfg, paymentRepo, sellerCredRepo)
	orderSvc := services.NewOrderService(orderRepo, productRepo, businessRepo, notifSvc, qrSvc, paymentSvc)

	// ── Handlers HTTP (adaptadores de entrada) ──────────────
	h := infraHTTP.Handlers{
		User:     handler.NewUserHandler(userSvc, notifSvc),
		Business: handler.NewBusinessHandler(businessSvc),
		Product:  handler.NewProductHandler(productSvc),
		Order:    handler.NewOrderHandler(orderSvc),
		Payment:  handler.NewPaymentHandler(paymentSvc),
	}

	// ── Job: cancelar pedidos expirados cada 5 min ──────────
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ids, err := orderRepo.(interface {
				CancelExpired(context.Context) ([]string, error)
			}).CancelExpired(context.Background())
			if err != nil {
				log.Printf("cancelExpired: %v", err)
			} else if len(ids) > 0 {
				log.Printf("cancelExpired: %d orders cancelled, processing refunds...", len(ids))
				for _, id := range ids {
					if err := paymentSvc.CancelPayment(context.Background(), id); err != nil {
						log.Printf("cancelExpired refund orderID=%s: %v", id, err)
					}
				}
			}
		}
	}()

	// ── Servidor HTTP ────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      infraHTTP.NewRouter(h, jwtSvc),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("🚀  POS API listening on %s\n", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
