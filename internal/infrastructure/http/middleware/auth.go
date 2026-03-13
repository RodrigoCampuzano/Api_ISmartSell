package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/jwt"
	"github.com/RodrigoCampuzano/Api_ISmartSell/pkg/response"
)

type contextKey string

const (
	CtxUserID contextKey = "userID"
	CtxRole   contextKey = "role"
)

// Auth valida el Bearer token y pone userID + role en el contexto.
func Auth(jwtSvc *jwt.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := jwtSvc.Parse(strings.TrimPrefix(auth, "Bearer "))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole devuelve 403 si el rol no coincide.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(CtxRole) != role {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(CtxUserID).(string)
	return v
}
