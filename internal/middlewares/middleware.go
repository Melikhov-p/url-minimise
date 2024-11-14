package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/compressor"
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

	Middleware func(http.HandlerFunc) http.HandlerFunc
)

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

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

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
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
						logger.Log.Error("error closing compressWriter", zap.Error(err))
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
					logger.Log.Error("error closing compressReader", zap.Error(err))
				}
			}()
		}

		h.ServeHTTP(ow, r)
	}

	return comp
}
