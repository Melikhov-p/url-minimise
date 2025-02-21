package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/contextkeys"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestWithLogging(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	store, err := repository.NewStorage(config.NewConfig(log, true), log)
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)

	middleware := Middleware{
		Logger:  log,
		Storage: store,
		Cfg:     cfg,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.WithLogging(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test", rr.Body.String())
}

func TestWithAuth(t *testing.T) {
	log, err := logger.BuildLogger("ERROR")
	assert.NoError(t, err)
	store, err := repository.NewStorage(config.NewConfig(log, true), log)
	assert.NoError(t, err)
	cfg := config.NewConfig(log, true)

	middleware := Middleware{
		Logger:  log,
		Storage: store,
		Cfg:     cfg,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(contextkeys.ContextUserKey).(*models.User)
		if user.ID == 1 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "Token", Value: "invalid token"})
	rr := httptest.NewRecorder()

	middleware.WithAuth(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
