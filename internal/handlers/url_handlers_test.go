package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/auth"
	"github.com/Melikhov-p/url-minimise/internal/config"
	loggerBuilder "github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/middlewares"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*config.Config, *zap.Logger) {
	logger, err := loggerBuilder.BuildLogger("DEBUG")
	assert.NoError(t, err)
	cfg := config.NewConfig(logger, true)

	return cfg, logger
}

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
			body:                createRandomURL(),
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

	router := chi.NewRouter()

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			CreateShortURL(w, r, cfg, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()
	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			r := resty.New().R()
			r.URL = srv.URL + "/"
			r.Method = test.method
			r.Body = strings.NewReader(test.body)
			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
			assert.Equal(t, test.expectedContentType, resp.Header().Get(`Content-Type`))
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

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.shortURL, http.NoBody)
			w := httptest.NewRecorder()

			GetFullURL(w, request, cfg, storage, logger)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}

func TestHappyPath(t *testing.T) {
	router := chi.NewRouter()

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			CreateShortURL(w, r, cfg, storage, logger)
		})
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		GetFullURL(w, r, cfg, storage, logger)
	})

	srv := httptest.NewServer(router)
	defer srv.Close()
	urlTest := createRandomURL()

	testCreate := struct {
		method       string
		fullURL      string
		expectedCode int
	}{
		method:       http.MethodPost,
		fullURL:      urlTest,
		expectedCode: http.StatusCreated}
	testGet := struct {
		method       string
		shortURL     string
		fullURL      string
		expectedCode int
	}{
		method:       http.MethodGet,
		shortURL:     "",
		fullURL:      urlTest,
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

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/api/shorten",
		func(w http.ResponseWriter, r *http.Request) {
			APICreateShortURL(w, r, cfg, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()
	urlTest := createRandomURL()
	requestBody := fmt.Sprintf(`{"url":"%s"}`, urlTest)

	testCases := []struct {
		name         string
		request      string
		method       string
		expectedCode int
	}{
		{
			name:         "APIHappyTest",
			request:      requestBody,
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "APIHappyTest",
			request:      requestBody,
			method:       http.MethodPost,
			expectedCode: http.StatusConflict,
		},
		{
			name:         "APIMethodNotAllowedTest",
			request:      requestBody,
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

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/api/shorten",
		func(w http.ResponseWriter, r *http.Request) {
			APICreateShortURL(w, r, cfg, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		urlTest := createRandomURL()
		requestBody := fmt.Sprintf(`{"url":"%s"}`, urlTest)
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
		urlTest := createRandomURL()
		requestBody := fmt.Sprintf(`{"url":"%s1"}`, urlTest)
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

func TestAPICreateBatchShortURL(t *testing.T) {
	router := chi.NewRouter()

	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/api/shorten/batch",
		func(w http.ResponseWriter, r *http.Request) {
			APICreateBatchURLs(w, r, cfg, storage, logger)
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
			name: "APIBatchHappyTest",
			request: fmt.Sprintf(`[
								{
									"correlation_id": "1",
									"original_url": "%s"
								},
								{
									"correlation_id": "2",
									"original_url": "%s"
								},
								{
									"correlation_id": "3",
									"original_url": "%s"
								}
							] `, createRandomURL(), createRandomURL(), createRandomURL()),
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "APIBatchMethodNotAllowedTest",
			request:      `{"url":"https://practicum.yandex.ru"}`,
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "APIBatchBadRequest",
			request:      `{"url":"https://practicum.yandex.ru"}`,
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			request := resty.New().R()
			request.URL = srv.URL + "/api/shorten/batch"
			request.Method = test.method

			request.SetHeader("Content-Type", "application/json")
			request.SetBody(test.request)

			resp, err := request.Send()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}

func TestAPIMarkAsDeletedURLs(t *testing.T) {
	router := chi.NewRouter()
	randURLsAmount := 120
	cfg, logger := setupTest(t)
	storage, err := repository.NewStorage(cfg, logger)
	assert.NoError(t, err)
	middleware := middlewares.Middleware{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
	router.Use(
		middleware.WithAuth,
		middleware.WithLogging,
		middleware.GzipMiddleware,
	)

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			CreateShortURL(w, r, cfg, storage, logger)
		})
	router.Delete("/api/user/urls",
		func(w http.ResponseWriter, r *http.Request) {
			APIMarkAsDeletedURLs(w, r, cfg, storage, logger)
		})
	router.Get("/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			GetFullURL(w, r, cfg, storage, logger)
		})

	srv := httptest.NewServer(router)
	defer srv.Close()

	newURLsForDelete := make([]string, 0, randURLsAmount)
	for i := 0; i < randURLsAmount; i++ {
		randURL := createRandomURL()
		newURLsForDelete = append(newURLsForDelete, randURL)
	}

	var shortURLs []string
	token, err := auth.BuildJWTString(999, cfg.SecretKey, 24*time.Hour)
	assert.NoError(t, err)

	logger.Debug("URLS", zap.Any("URLS", newURLsForDelete))
	for _, origURL := range newURLsForDelete {
		t.Run("create url", func(t *testing.T) {
			request := resty.New().R()
			request.URL = srv.URL + "/"
			request.Method = http.MethodPost

			request.SetHeader("Content-Type", "text/plain")
			request.SetBody(origURL)
			request.SetCookie(&http.Cookie{Name: "Token", Value: token})
			resp, err := request.Send()
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode())

			shortURLs = append(shortURLs, strings.Split(string(resp.Body()), "/")[3])
		})
	}

	delRequest := resty.New().R()
	delRequest.Method = http.MethodDelete
	delRequest.URL = srv.URL + "/api/user/urls"
	delRequest.SetCookie(&http.Cookie{Name: "Token", Value: token})
	delRequest.SetBody(shortURLs)

	delResp, err := delRequest.Send()
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, delResp.StatusCode())

	checkClient := resty.New()
	checkClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	checkRequest := checkClient.R()
	checkRequest.Method = http.MethodGet

	triesCount := 5 // Количество попыток если вдруг запросим урл который ещё не удалился
	for _, shortURL := range shortURLs {
		if triesCount == 0 {
			logger.Debug("tries for check URLs delete is 0")
			return
		}
		checkRequest.URL = srv.URL + "/" + shortURL
		checkResp, err := checkRequest.Send()
		if err != nil {
			time.Sleep(time.Second)
			logger.Debug("error checking deleted url", zap.Error(err))
			triesCount--
			continue
		}
		logger.Debug("url is gone", zap.String("URL", shortURL))
		assert.Equal(t, http.StatusGone, checkResp.StatusCode())
	}

}

// randomString генерирует случайную строку заданной длины
func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// createRandomURL генерирует случайный URL
func createRandomURL() string {
	scheme := "https"
	host := randomString(10) + ".example." + randomString(3)
	path := "/" + randomString(5)
	time.Sleep(5 * time.Millisecond)

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	return u.String()
}
