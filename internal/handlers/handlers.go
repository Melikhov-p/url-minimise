package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/logger"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CreateShortURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, storage *models.Storage) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fullURL, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		logger.Log.Error("error read body from text", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newURL, err := models.NewStorageURL(string(fullURL), storage)
	if err != nil {
		logger.Log.Error("error creating short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = storage.AddURL(newURL); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `text/plain`)
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintf(w, `%s%s`, cfg.ResultAddr+"/", newURL.ShortURL)

	if err != nil {
		logger.Log.Error("error writing body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetFullURL(w http.ResponseWriter, r *http.Request, storage *models.Storage) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := chi.URLParam(r, "id")

	matchURL := storage.GetFullURL(id)
	if matchURL == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set(`Location`, matchURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func APICreateShortURL(w http.ResponseWriter, r *http.Request, cfg *config.Config, storage *models.Storage) {
	if r.Method != http.MethodPost {
		logger.Log.Info("wrong method used", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	logger.Log.Debug("start decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		logger.Log.Error("error decoding request json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newURL, err := models.NewStorageURL(req.URL, storage)
	if err != nil {
		logger.Log.Error("error creating short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = storage.AddURL(newURL); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("start encoding response")
	res := models.Response{
		ResultURL: cfg.ResultAddr + "/" + newURL.ShortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err = enc.Encode(res); err != nil && !errors.Is(err, io.EOF) {
		logger.Log.Error("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
