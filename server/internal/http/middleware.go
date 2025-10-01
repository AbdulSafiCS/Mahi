package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKeyUserID struct{}

func (s *Server) authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing_bearer", nil)
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := s.jwt.Parse(raw)
		if err != nil || claims == nil || claims.UserID == "" {
			if err != nil && err == jwt.ErrTokenExpired {
				writeErr(w, http.StatusUnauthorized, "token_expired", nil)
				return
			}
			writeErr(w, http.StatusUnauthorized, "token_invalid", nil)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID{}, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
