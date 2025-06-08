package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type logRecord struct {
	statusCode int
	http.ResponseWriter
}

func (lr *logRecord) WriteHeader(code int) {
	lr.statusCode = code
	lr.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware returns an HTTP middleware that logs incoming requests and responses.
// It generates or propagates a correlation ID (X-Correlation-ID) for tracing requests,
// logs the request start with method and URI, then logs the response status and duration.
func LoggingMiddleware(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid := r.Header.Get("X-Correlation-ID")
			if cid == "" {
				cid = uuid.New().String()
			}
			w.Header().Set("X-Correlation-ID", cid)

			start := time.Now()
			logger.Info("start request",
				zap.String("correlation_id", cid),
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
			)

			lr := &logRecord{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(lr, r)

			duration := time.Since(start)
			logger.Info("end request",
				zap.String("correlation_id", cid),
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", lr.statusCode),
				zap.Duration("duration", duration),
			)
		})
	}
}
