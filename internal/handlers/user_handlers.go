package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"go.uber.org/zap"
)

func GetUserURLs(
	w http.ResponseWriter,
	r *http.Request,
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.Info("unresolved method for get user urls")
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		logger.Error("error priveniye tipa user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
			ShortURL:    cfg.ResultAddr + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}
	w.WriteHeader(http.StatusOK)

	if err = enc.Encode(&res.UserURLs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("error encoding user's urls response", zap.Error(err))
	}
}
