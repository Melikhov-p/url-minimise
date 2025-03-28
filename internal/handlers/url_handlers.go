package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/contextkeys"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
	storagePkg "github.com/Melikhov-p/url-minimise/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	errGetContextUser = errors.New("error getting user from context")
)

// CreateShortURL создание нового URL.
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

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		logger.Error(errGetContextUser.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, storage, cfg)
		if err != nil {
			logger.Error("error adding new user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: user.Service.Token,
	})

	newURL, err := service.AddURL(ctx, storage, logger, string(originalURL), cfg, user.ID)
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
	logger.Debug("add new URL from /",
		zap.String("OriginalURL", newURL.OriginalURL),
		zap.String("ShortURL", newURL.ShortURL),
		zap.Int("USER_ID", user.ID))
	_, err = fmt.Fprintf(w, `%s%s`, cfg.ResultAddr+"/", newURL.ShortURL)

	if err != nil {
		logger.Error("error writing body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GetFullURL Получение полного URL.
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

	matchURL, err := storage.GetURL(ctx, id)
	if err != nil {
		logger.Info("not found full URL by short", zap.String("shortURL", id), zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if matchURL.DeletedFlag {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set(`Location`, matchURL.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// APICreateShortURL создание нового URL через json.
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

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		logger.Error(errGetContextUser.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var err error
	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, storage, cfg)
		if err != nil {
			logger.Error("error adding new user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: user.Service.Token,
	})

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

	newURL, err := service.AddURL(ctx, storage, logger, req.URL, cfg, user.ID)
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
	logger.Debug("add new URL from /api/shorten",
		zap.String("OriginalURL", newURL.OriginalURL),
		zap.String("ShortURL", newURL.ShortURL))
	if err = enc.Encode(res); err != nil && !errors.Is(err, io.EOF) {
		logger.Error("error encoding response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// APICreateBatchURLs Создание пачки новых URL.
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

	ctx := r.Context()
	dec := json.NewDecoder(r.Body)
	var req models.BatchRequest
	if err := dec.Decode(&req.BatchURLs); err != nil {
		logger.Error("error decoding request to request model", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		logger.Error(errGetContextUser.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var err error
	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, storage, cfg)
		if err != nil {
			logger.Error("error adding new user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: user.Service.Token,
	})

	originalURLs := make([]string, 0, len(req.BatchURLs))
	for _, url := range req.BatchURLs {
		originalURLs = append(originalURLs, url.OriginalURL)
	}

	newURLs, err := repository.NewStorageMultiURL(r.Context(), originalURLs, storage, cfg, user.ID)
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
	logger.Debug("add new URLs from /api/shorten/batch")

	if err = enc.Encode(&res.BatchURLs); err != nil {
		logger.Error("error encoding batch response to json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// APIMarkAsDeletedURLs пометить URL на удаление.
func APIMarkAsDeletedURLs(
	w http.ResponseWriter,
	r *http.Request,
	_ *config.Config,
	storage repository.Storage,
	logger *zap.Logger,
) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		logger.Error("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !user.Service.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&shortURLs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("error decoding delete request", zap.Error(err))
		return
	}

	go func() {
		err := storage.AddDeleteTask(shortURLs, user.ID)
		if err != nil {
			logger.Error("error adding new delete task in storage", zap.Error(err))
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}

func GetServiceStats(
	w http.ResponseWriter,
	r *http.Request,
	cfg *config.Config,
	storage repository.Storage,
	logger *zap.Logger,
) {
	ctx := r.Context()

	var (
		usrIPHeader string
		usrIP       net.IP
		trustedNet  *net.IPNet
		err         error
		usersCount  int
		urlsCount   int
	)

	if usrIPHeader = r.Header.Get("X-Real-IP"); usrIPHeader == "" {
		logger.Error("empty X-Real-IP header in stats request")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	usrIP = net.ParseIP(usrIPHeader)
	if usrIP == nil {
		logger.Error("error parsing IP from header", zap.String("IP header", usrIPHeader))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, trustedNet, err = net.ParseCIDR(cfg.TrustedSubNet)
	if err != nil {
		logger.Error("error parsing CIDR from config", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !trustedNet.Contains(usrIP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	usersCount, err = storage.GetUsersCount(ctx)
	if err != nil {
		logger.Error("error getting users count", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	urlsCount, err = storage.GetURLsCount(ctx)
	if err != nil {
		logger.Error("error getting urls count", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sr := models.StatsResponse{
		URLs:  urlsCount,
		Users: usersCount,
	}

	enc := json.NewEncoder(w)
	if err = enc.Encode(sr); err != nil {
		logger.Error("error encoding json to stats request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PingDatabase проверка соединения с базой данных.
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
