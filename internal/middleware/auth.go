package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/yohnnn/booking_service/internal/handler/response"
	"github.com/yohnnn/booking_service/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(authService service.Auth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteErrorResponse(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "missing authorization header")
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				response.WriteErrorResponse(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "invalid authorization header")
				return
			}

			token := headerParts[1]
			userID, err := authService.ParseToken(token)
			if err != nil {
				response.WriteErrorResponse(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
