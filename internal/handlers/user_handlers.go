package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
	"go.uber.org/zap"
)

func GetUserURLs(
	w http.ResponseWriter,
	r *http.Request,
	_ *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.Info("unresolved method for get user urls")
		return
	}

	tokenCookie, err := r.Cookie("Token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("error getting token cookie", zap.Error(err))
		return
	}
	if errors.Is(err, http.ErrNoCookie) {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("unauthorized request")
		return
	}

	user, err := service.AuthUserByToken(tokenCookie.Value, storage, logger)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Error("error authenticate user", zap.Error(err))
		return
	}

	ctx := r.Context()
	urls, err := storage.GetURLsByUserID(ctx, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("error getting user's urls", zap.Error(err))
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")

	var res models.UserURLsResponse
	for _, url := range urls {
		res.UserURLs = append(res.UserURLs, &models.UserURL{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}
	w.WriteHeader(http.StatusOK)

	if err = enc.Encode(&res.UserURLs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("error encoding user's urls response", zap.Error(err))
	}

}
