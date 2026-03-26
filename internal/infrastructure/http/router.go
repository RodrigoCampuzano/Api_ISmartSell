package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/handler"
	"github.com/RodrigoCampuzano/Api_ISmartSell/internal/infrastructure/http/middleware"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/jwt"
)

type Handlers struct {
	User     *handler.UserHandler
	Business *handler.BusinessHandler
	Product  *handler.ProductHandler
	Order    *handler.OrderHandler
}

func NewRouter(h Handlers, jwtSvc *jwt.Service) http.Handler {
	r := chi.NewRouter()

	// Middlewares globales
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)

	r.Route("/api/v1", func(r chi.Router) {

		// ── Auth (público) ──────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", h.User.Register)
			r.Post("/login", h.User.Login)
		})

		// ── Rutas protegidas ────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))

			// Perfil propio
			r.Get("/users/me", h.User.Me)

			// Negocios (públicos con auth)
			r.Get("/businesses", h.Business.ListNearby)
			r.Get("/businesses/{id}", h.Business.Get)

			// Negocios — solo vendedor
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("seller"))
				r.Post("/businesses", h.Business.Create)
				r.Delete("/businesses/{id}", h.Business.Delete)
				r.Get("/businesses/mine", h.Business.ListMine)
				r.Post("/businesses/{id}/delivery-points", h.Business.AddDeliveryPoint)

				// Productos: crear/editar/borrar
				r.Post("/businesses/{businessId}/products", h.Product.Create)
				r.Put("/products/{id}", h.Product.Update)
				r.Delete("/products/{id}", h.Product.Delete)

				// Órdenes del negocio
				r.Get("/businesses/{businessId}/orders", h.Order.ListByBusiness)
				r.Post("/orders/{id}/ready", h.Order.Ready)
				r.Post("/orders/scan", h.Order.ScanQR)
			})

			// Productos — lectura pública (autenticado)
			r.Get("/businesses/{businessId}/products", h.Product.ListByBusiness)
			r.Get("/products/{id}", h.Product.Get)

			// Órdenes — comprador
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("buyer"))
				r.Post("/orders", h.Order.Create)
				r.Get("/orders/my", h.Order.ListMine)
				r.Post("/orders/{id}/cancel", h.Order.Cancel)
			})

			// Ver orden (comprador o vendedor del negocio)
			r.Get("/orders/{id}", h.Order.Get)
		})
	})

	return r
}
