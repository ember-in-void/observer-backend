// internal/shared/http/middleware/logging.go
package middleware

import (
	"net/http"
	"time"

	"steam-observer/internal/shared/logger"
)

// responseWriter обёртка для захвата статус-кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware логирует все HTTP запросы
func Logging(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Оборачиваем ResponseWriter для захвата статус-кода
			wrapped := newResponseWriter(w)

			// Вызываем следующий handler
			next.ServeHTTP(wrapped, r)

			// Логируем после выполнения
			duration := time.Since(start)

			logEntry := log.WithFields(map[string]any{
				"method":   r.Method,
				"path":     r.URL.Path,
				"status":   wrapped.statusCode,
				"duration": duration.String(),
				"ip":       r.RemoteAddr,
			})

			// Разные уровни логирования в зависимости от статус-кода
			switch {
			case wrapped.statusCode >= 500:
				logEntry.Error("server error")
			case wrapped.statusCode >= 400:
				logEntry.Warn("client error")
			default:
				logEntry.Info("request completed")
			}
		})
	}
}
