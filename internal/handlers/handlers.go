package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
	storagePkg "github.com/Melikhov-p/url-minimise/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CreateShortURL(
	w http.ResponseWriter,
	r *http.Request,
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("error reading body from text/plain", zap.Error(err))
	}

	defer func() {
		_ = r.Body.Close()
	}()

	ctx := r.Context()

	newURL, err := service.AddURL(ctx, storage, logger, string(originalURL), cfg)
	if err != nil {
		if errors.Is(err, storagePkg.ErrOriginalURLExist) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error adding new URL", zap.Error(err))
			return
		}
	}

	if saver, ok := storage.(repository.StorageSaver); ok {
		if err = saver.Save(newURL); err != nil {
			logger.Error("error saving new URL %v", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set(`Content-Type`, `text/plain`)
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintf(w, `%s%s`, cfg.ResultAddr+"/", newURL.ShortURL)

	if err != nil {
		logger.Error("error writing body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetFullURL(
	w http.ResponseWriter,
	r *http.Request,
	_ *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	id := chi.URLParam(r, "id")

	matchURL, err := storage.GetFullURL(ctx, id)
	if err != nil {
		logger.Info("not found full URL by short", zap.String("shortURL", id))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set(`Location`, matchURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func APICreateShortURL(
	w http.ResponseWriter,
	r *http.Request,
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	logger.Debug("start decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		logger.Error("error decoding request json", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")

	newURL, err := service.AddURL(ctx, storage, logger, req.URL, cfg)
	if err != nil {
		if errors.Is(err, storagePkg.ErrOriginalURLExist) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error adding new URL", zap.Error(err))
			return
		}
	}

	if saver, ok := storage.(repository.StorageSaver); ok {
		if err = saver.Save(newURL); err != nil {
			logger.Error("error saving new URL %v", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	res := models.Response{
		ResultURL: cfg.ResultAddr + "/" + newURL.ShortURL,
	}
	w.WriteHeader(http.StatusCreated)
	if err = enc.Encode(res); err != nil && !errors.Is(err, io.EOF) {
		logger.Error("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func APICreateBatchURLs(
	w http.ResponseWriter,
	r *http.Request,
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger,
) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(r.Body)
	var req models.BatchRequest
	if err := dec.Decode(&req.BatchURLs); err != nil {
		logger.Error("error decoding request to request model", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURLs := make([]string, 0, len(req.BatchURLs))
	for _, url := range req.BatchURLs {
		originalURLs = append(originalURLs, url.OriginalURL)
	}

	newURLs, err := repository.NewStorageMultiURL(r.Context(), originalURLs, storage, cfg)
	if err != nil {
		logger.Error("error getting new storage for multi urls", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res models.BatchResponse
	for i, url := range newURLs {
		res.BatchURLs = append(res.BatchURLs, models.BatchURLResponse{
			ShortURL:      cfg.ResultAddr + "/" + url.ShortURL,
			CorrelationID: req.BatchURLs[i].CorrelationID,
		})
	}

	if err = storage.AddURLs(r.Context(), newURLs); err != nil {
		logger.Error("error adding new urls to database", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = enc.Encode(&res.BatchURLs); err != nil {
		logger.Error("error encoding batch response to json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func PingDatabase(
	w http.ResponseWriter,
	r *http.Request,
	_ *config.Config,
	storage repository.Storage,
	logger *zap.Logger,
) {
	ctx := r.Context()

	if err := storage.Ping(ctx); err != nil {
		logger.Error("database is unavailable from ping method", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
