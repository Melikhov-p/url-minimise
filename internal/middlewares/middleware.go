package middlewares

import (
	"net/http"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFunc := func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggerResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(startTime)

		logger.Log.Info(
			"",
			zap.String("URI", r.RequestURI),
			zap.String("METHOD", r.Method),
			zap.Duration("DURATION", duration),
			zap.Int("SIZE", responseData.size),
			zap.Int("STATUS", responseData.status),
		)
	}

	return logFunc
}

func (r *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
