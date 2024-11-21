package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	compress "github.com/Melikhov-p/url-minimise/internal/compressor"
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

	Middleware struct {
		Logger *zap.Logger
	}
)

func (m *Middleware) WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		m.Logger.Info(
			"",
			zap.String("URI", r.RequestURI),
			zap.String("METHOD", r.Method),
			zap.Duration("DURATION", duration),
			zap.Int("SIZE", responseData.size),
			zap.Int("STATUS", responseData.status),
		)
	})
}

func (r *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if err != nil {
		return size, fmt.Errorf("error write from loggerResponseWriter %w", err)
	}
	return size, nil
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (m *Middleware) GzipMiddleware(h http.Handler) http.Handler {
	comp := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		content := w.Header().Get("Content-Type")

		if content == "application/json" || content == "text/html" {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				cw := compress.NewCompressWrite(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						m.Logger.Error("error closing compressWriter", zap.Error(err))
					}
				}()
			}
		}

		contentEncoding := r.Header.Get("Content-Encoding")

		if strings.Contains(contentEncoding, "gzip") {
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				if err = cr.Close(); err != nil {
					m.Logger.Error("error closing compressReader", zap.Error(err))
				}
			}()
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(comp)
}
