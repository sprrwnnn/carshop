package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (m *LoggingMiddleware) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var (
			start     = time.Now()
			requestID = uuid.NewString()
		)

		m.logger.Info("Started request",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.String("request_id", requestID),
		)

		wrappedWriter := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		defer func() {
			duration := time.Since(start)

			if err := recover(); err != nil {
				m.logger.Error("panic in request handler",
					zap.String("method", req.Method),
					zap.String("path", req.URL.Path),
					zap.Int("status_code", wrappedWriter.statusCode),
					zap.Any("error", err),
					zap.Duration("duration", duration),
					zap.String("request_id", requestID),
				)

				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			m.logger.Info("Completed request",
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status_code", wrappedWriter.statusCode),
				zap.Duration("duration", duration),
				zap.String("request_id", requestID),
			)
		}()

		next.ServeHTTP(wrappedWriter, req)
	})
}
