package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetUserURLs(t *testing.T) {
	router := chi.NewRouter()

	cfg, log := setupTest(t)
	storage, err := repository.NewStorage(cfg, log)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  log,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Get("/",
		func(w http.ResponseWriter, r *http.Request) {
			GetUserURLs(w, r, cfg, storage, log)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		name                string
		method              string
		expectedCode        int
		expectedContentType string
		body                string
	}{
		{
			name:                "Method Not Allowed",
			method:              http.MethodPost,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
			body:                ``,
		},
		{
			name:                "Method Not Allowed",
			method:              http.MethodDelete,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
			body:                ``,
		},
		{
			name:                "No Content",
			method:              http.MethodGet,
			expectedCode:        http.StatusNoContent,
			expectedContentType: ``,
			body:                ``,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := resty.New().R()
			request.URL = srv.URL + "/"
			request.Method = testCase.method

			request.SetHeader("Content-Type", "application/json")

			resp, err := request.Send()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedCode, resp.StatusCode())
		})
	}
}
