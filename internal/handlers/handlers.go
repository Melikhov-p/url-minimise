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
	fullURL, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err != nil {
		logger.Error("error read body from text", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx := r.Context()

	newURL, err := repository.NewStorageURL(ctx, string(fullURL), storage, cfg)
	if err != nil {
		logger.Error("error creating short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = storage.AddURL(ctx, newURL); err != nil {
		if errors.Is(err, storagePkg.ErrURLExist) {
			logger.Info("original URL already exist in storage", zap.String("OriginalURL", newURL.OriginalURL))

			if existShort, err := storage.GetShortURL(r.Context(), newURL.OriginalURL); err == nil {
				logger.Info("short url found in storage", zap.String("shortURL", existShort))
				w.WriteHeader(http.StatusConflict)
				if _, err = fmt.Fprintf(w, `%s%s`, cfg.ResultAddr+"/", existShort); err != nil {
					logger.Error("error write response to io.Writer", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			} else {
				logger.Error("short URL dont found for existing original URL",
					zap.String("Original", newURL.OriginalURL),
					zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		logger.Error("error adding new url", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
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
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.Info("unresolved method", zap.String("method", r.Method))
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
		logger.Info("wrong method used", zap.String("method", r.Method))
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

	newURL, err := repository.NewStorageURL(ctx, req.URL, storage, cfg)
	if err != nil {
		logger.Error("error creating short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	res := models.Response{
		ResultURL: cfg.ResultAddr + "/" + newURL.ShortURL,
	}

	if err = storage.AddURL(ctx, newURL); err != nil {
		if errors.Is(err, storagePkg.ErrURLExist) {
			logger.Info("original URL already exist in storage", zap.String("OriginalURL", newURL.OriginalURL))

			if existShort, err := storage.GetShortURL(r.Context(), newURL.OriginalURL); err == nil {
				logger.Info("short url found in storage", zap.String("shortURL", existShort))
				w.WriteHeader(http.StatusConflict)
				if encErr := enc.Encode(res); encErr != nil {
					logger.Error("error encoding response", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				logger.Error("short URL dont found for existing original URL",
					zap.String("Original", newURL.OriginalURL),
					zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		logger.Error("error adding new url", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if saver, ok := storage.(repository.StorageSaver); ok {
		if err = saver.Save(newURL); err != nil {
			logger.Error("error saving new URL %v", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	logger *zap.Logger) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	var req models.BatchRequest
	if err := dec.Decode(&req.BatchURLs); err != nil {
		logger.Error("error decoding request to request model", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
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
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if cfg.StorageMode != storagePkg.StorageInDatabase {
		logger.Error("ping database with wrong storage type", zap.Any("storage type", cfg.StorageMode))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx := r.Context()

	if err := storage.Ping(ctx); err != nil {
		logger.Error("database is unavailable from ping method", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
