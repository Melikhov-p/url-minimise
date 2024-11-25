package handlers

import (
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var CFG *config.Config = config.NewConfig()

func TestCreateShortURL(t *testing.T) {
	testCases := []struct {
		method              string
		expectedCode        int
		expectedContentType string
		body                string
	}{
		{
			method:              http.MethodPost,
			expectedCode:        http.StatusCreated,
			expectedContentType: `text/plain`,
			body:                `https://github.com/Melikhov-p/url-minimise/pull/5`,
		},
		{
			method:              http.MethodGet,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
			body:                ``,
		},
		{
			method:              http.MethodPut,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
			body:                ``,
		},
		{
			method:              http.MethodDelete,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
			body:                ``,
		},
	}

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()
			logger, err := loggerBuilder.BuildLogger("DEBUG")
			assert.NoError(t, err)
			storage, err := repository.NewStorage(CFG)
			assert.NoError(t, err)
			CreateShortURL(w, request, CFG, storage, logger)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}

func TestGetFullURL(t *testing.T) {
	testCases := []struct {
		method              string
		shortURL            string
		expectedCode        int
		expectedContentType string
	}{
		{
			method:              http.MethodPost,
			shortURL:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodPut,
			shortURL:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodDelete,
			shortURL:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodGet,
			shortURL:            `TEST`,
			expectedCode:        http.StatusNotFound,
			expectedContentType: ``,
		},
	}

	storage, err := repository.NewStorage(CFG)
	assert.NoError(t, err)
	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.shortURL, http.NoBody)
			w := httptest.NewRecorder()
			logger, err := loggerBuilder.BuildLogger("DEBUG")
			assert.NoError(t, err)
			GetFullURL(w, request, storage, logger)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}

func TestHappyPath(t *testing.T) {
	router := chi.NewRouter()

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	assert.NoError(t, err)
	storage, err := repository.NewStorage(CFG)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{Logger: logger}
	router.Use(
		middleware.WithLogging,
	)

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			CreateShortURL(w, r, CFG, storage, logger)
		})
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		GetFullURL(w, r, storage, logger)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCreate := struct {
		method       string
		fullURL      string
		expectedCode int
	}{
		method:       http.MethodPost,
		fullURL:      "https://github.com/Melikhov-p/url-minimise/pull/5",
		expectedCode: http.StatusCreated}
	testGet := struct {
		method       string
		shortURL     string
		fullURL      string
		expectedCode int
	}{
		method:       http.MethodGet,
		shortURL:     "",
		fullURL:      "https://github.com/Melikhov-p/url-minimise/pull/5",
		expectedCode: http.StatusTemporaryRedirect,
	}

	t.Run("HappyPathTest", func(t *testing.T) {
		createRequest := resty.New().R()
		createRequest.URL = srv.URL + "/"
		createRequest.Method = testCreate.method
		createRequest.Body = strings.NewReader(testCreate.fullURL)

		createResponse, err := createRequest.Send()
		assert.NoError(t, err)
		assert.Equal(t, testCreate.expectedCode, createResponse.StatusCode())

		testGet.shortURL = strings.Split(string(createResponse.Body()), "/")[3]

		getRequest := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R()
		getRequest.Method = testGet.method
		getRequest.URL = srv.URL + "/" + testGet.shortURL

		getResponse, err := getRequest.Send()
		log.Printf("%v", getResponse)
		assert.Error(t, err)

		assert.Equal(t, testGet.expectedCode, getResponse.StatusCode())
		assert.Equal(t, testGet.fullURL, getResponse.Header().Get("Location"))
	})
}

func TestAPICreateShortURL(t *testing.T) {
	router := chi.NewRouter()

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	assert.NoError(t, err)
	storage, err := repository.NewStorage(CFG)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{Logger: logger}
	router.Use(
		middleware.WithLogging,
	)

	router.Post("/api/shorten",
		func(w http.ResponseWriter, r *http.Request) {
			APICreateShortURL(w, r, CFG, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()

	testCases := []struct {
		name         string
		request      string
		method       string
		expectedCode int
	}{
		{
			name:         "APIHappyTest",
			request:      `{"url":"https://practicum.yandex.ru"}`,
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "APIMethodNotAllowedTest",
			request:      `{"url":"https://practicum.yandex.ru"}`,
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.URL = srv.URL + "/api/shorten"
			request.Method = test.method

			request.SetHeader("Content-Type", "application/json")
			request.SetBody(test.request)

			resp, err := request.Send()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}

func TestCompressor(t *testing.T) {
	router := chi.NewRouter()

	logger, err := loggerBuilder.BuildLogger("DEBUG")
	assert.NoError(t, err)
	storage, err := repository.NewStorage(CFG)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{Logger: logger}
	router.Use(
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/api/shorten",
		func(w http.ResponseWriter, r *http.Request) {
			APICreateShortURL(w, r, CFG, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()

	requestBody := `{"url": "https://practicum.yandex.ru/"}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		request := resty.New().R()
		request.URL = srv.URL + "/api/shorten"
		request.Method = http.MethodPost
		request.Body = buf
		request.Header.Set("Content-Encoding", "gzip")

		response, reqErr := request.Send()

		assert.NoError(t, reqErr)
		assert.Equal(t, http.StatusCreated, response.StatusCode())
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		request := resty.New().R()
		request.URL = srv.URL + "/api/shorten"
		request.Method = http.MethodPost
		request.Body = buf
		request.Header.Add("Accept-Encoding", "gzip")

		resp, err := request.Send()

		require.NoError(t, err)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode())
	})
}
