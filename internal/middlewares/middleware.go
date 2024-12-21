package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	compress "github.com/Melikhov-p/url-minimise/internal/compressor"
	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/contextkeys"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
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
		Logger  *zap.Logger
		Storage repository.Storage
		Cfg     *config.Config
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

		user, ok := r.Context().Value("user").(*models.User)
		if !ok {
			m.Logger.Error("error getting user from context in logging middleware")
			user = repository.NewEmptyUser()
		}

		m.Logger.Info(
			"",
			zap.String("URI", r.RequestURI),
			zap.String("METHOD", r.Method),
			zap.Duration("DURATION", duration),
			zap.Int("SIZE", responseData.size),
			zap.Int("STATUS", responseData.status),
			zap.Int("USER_ID", user.ID),
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
				m.Logger.Error("error getting compress reader", zap.Error(err))
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

func (m *Middleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("Token")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusBadRequest)
			m.Logger.Error("can not read cookie from request", zap.Error(err))
			return
		}

		var user *models.User
		if !errors.Is(err, http.ErrNoCookie) {
			token := tokenCookie.Value
			user, err = service.AuthUserByToken(token, m.Storage, m.Logger, m.Cfg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				m.Logger.Error("error authorizing user", zap.Error(err))
				return
			}
		} else {
			user, err = service.AddNewUser(r.Context(), m.Storage, m.Cfg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				m.Logger.Error("error creating new user", zap.Error(err))
			}
		}

		ctxWithUser := context.WithValue(r.Context(), contextkeys.ContextUserKey, user)
		h.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
