// internal/shared/http/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"steam-observer/internal/modules/auth/ports/out_ports"
)

// локальный тип для ключа в контексте
type ctxKey string

const userIDKey ctxKey = "userID"

// Хелпер, чтобы хендлеры доставали userID из контекста
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(string)
	return id, ok
}

// Auth возвращает функцию-обёртку, которую можно применить к любому http.Handler.
func Auth(tokenProvider out_ports.TokenProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Достать Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing or invalid authorization header"}`))
				return
			}

			rawToken := strings.TrimPrefix(authHeader, "Bearer ")

			// 2. Распарсить токен и получить userID (через TokenProvider.ParseAccessToken)
			userID, _, err := tokenProvider.ParseAccessToken(r.Context(), rawToken)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid token"}`))
				return
			}

			// 3. Положить userID в context и вызвать следующий handler
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
