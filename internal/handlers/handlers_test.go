package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

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
			CreateShortURL(w, request, config.NewConfig())

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

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.shortURL, http.NoBody)
			w := httptest.NewRecorder()
			GetFullURL(w, request)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}

func TestHappyPath(t *testing.T) {
	router := chi.NewRouter()

	router.Post("/",
		func(w http.ResponseWriter, r *http.Request) {
			CreateShortURL(w, r, config.NewConfig())
		})
	router.Get("/{id}", GetFullURL)

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
