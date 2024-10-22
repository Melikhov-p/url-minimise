package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
	testCases := []struct {
		method              string
		expectedCode        int
		expectedContentType string
	}{
		{
			method:              http.MethodPost,
			expectedCode:        http.StatusCreated,
			expectedContentType: `text/plain`,
		},
		{
			method:              http.MethodGet,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodPut,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodDelete,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
	}

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/", nil)
			w := httptest.NewRecorder()
			CreateShortURL(w, request)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}

func TestGetFullURL(t *testing.T) {
	testCases := []struct {
		method              string
		shortUrl            string
		expectedCode        int
		expectedContentType string
	}{
		{
			method:              http.MethodPost,
			shortUrl:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodPut,
			shortUrl:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodDelete,
			shortUrl:            `TEST`,
			expectedCode:        http.StatusMethodNotAllowed,
			expectedContentType: ``,
		},
		{
			method:              http.MethodGet,
			shortUrl:            `TEST`,
			expectedCode:        http.StatusNotFound,
			expectedContentType: ``,
		},
	}

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.shortUrl, nil)
			w := httptest.NewRecorder()
			GetFullURL(w, request)

			assert.Equal(t, test.expectedCode, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get(`Content-Type`))
		})
	}
}
