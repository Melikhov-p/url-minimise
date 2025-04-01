package grpc

import (
	"context"
	"errors"
	"net"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/contextkeys"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
	storagePkg "github.com/Melikhov-p/url-minimise/internal/storage"
	"github.com/Melikhov-p/url-minimise/protos/gen/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Shortener gRPC сервис shortener'а.
type Shortener struct {
	proto.UnimplementedShortenerServer
	log   *zap.Logger
	cfg   *config.Config
	store repository.Storage
}

// NewShortenerService создаёт новый сервис shortener'а.
func NewShortenerService(log *zap.Logger, cfg *config.Config, store repository.Storage) *Shortener {
	return &Shortener{
		log:   log,
		cfg:   cfg,
		store: store,
	}
}

// CreateURL создает новый короткий URL.
func (s *Shortener) CreateURL(ctx context.Context, in *proto.CreateURLRequest) (*proto.CreateURLResponse, error) {
	var (
		res  proto.CreateURLResponse
		user *models.User
		err  error
	)

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		return nil, status.Error(codes.Internal, "error getting user")
	}

	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, s.store, s.cfg)
		if err != nil {
			s.log.Error("error adding new user", zap.Error(err))
			return nil, status.Error(codes.Internal, "error adding new user.")
		}
	}

	err = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"authorization": user.Service.Token,
	}))
	if err != nil {
		s.log.Error("error adding auth metadata in grpc response", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "error returning token")
	}

	newURL, err := service.AddURL(ctx, s.store, s.log, in.GetOriginalUrl(), s.cfg, user.ID)
	if err != nil {
		if errors.Is(err, storagePkg.ErrOriginalURLExist) {
			return nil, status.Error(codes.AlreadyExists, "original URL already exist.")
		} else {
			return nil, status.Error(codes.Internal, "")
		}
	}

	if saver, ok := s.store.(repository.StorageSaver); ok {
		if err = saver.Save(newURL); err != nil {
			s.log.Error("error saving new url", zap.Error(err), zap.String("original_url", in.GetOriginalUrl()))
			return nil, status.Error(codes.Internal, "")
		}
	}

	res.ShortUrl = newURL.ShortURL
	return &res, nil
}

// GetFullURL возвращает оригинальный URL.
func (s *Shortener) GetFullURL(ctx context.Context, in *proto.GetFullURLRequest) (*proto.GetFullURLResponse, error) {
	var (
		res      proto.GetFullURLResponse
		matchURL *models.StorageURL
		err      error
	)

	matchURL, err = s.store.GetURL(ctx, in.GetShortUrl())
	if err != nil {
		s.log.Error("error finding original URL", zap.String("short", in.GetShortUrl()))
		return nil, status.Error(codes.NotFound, "original url not found.")
	}

	res.OriginalUrl = matchURL.OriginalURL

	return &res, nil
}

// CreateBatchURLs создает пачку новых URL.
func (s *Shortener) CreateBatchURLs(
	ctx context.Context,
	in *proto.CreateBatchURLRequest,
) (*proto.CreateBatchURLResponse, error) {
	var (
		res proto.CreateBatchURLResponse
		err error
	)

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		return nil, status.Error(codes.Internal, "error getting user.")
	}
	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, s.store, s.cfg)
		if err != nil {
			s.log.Error("error adding new user", zap.Error(err))
			return nil, status.Error(codes.Internal, "error adding new user.")
		}
	}

	err = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"authorization": user.Service.Token,
	}))
	if err != nil {
		s.log.Error("error adding auth metadata in grpc response", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "error returning token")
	}

	originalURLs := make([]string, 0, len(in.GetBatchUrls()))
	for _, url := range in.GetBatchUrls() {
		originalURLs = append(originalURLs, url.GetOriginalUrl())
	}

	newURLs, err := repository.NewStorageMultiURL(ctx, originalURLs, s.store, s.cfg, user.ID)
	if err != nil {
		s.log.Error("error getting new storage for multi urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	if err = s.store.AddURLs(ctx, newURLs); err != nil {
		s.log.Error("error adding new urls to database", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	for i, url := range newURLs {
		res.BatchUrls = append(res.BatchUrls, &proto.BatchResponseURL{
			ShortUrl:      s.cfg.ResultAddr + "/" + url.ShortURL,
			CorrelationId: in.GetBatchUrls()[i].GetCorrelationId(),
		})
	}

	return &res, nil
}

// MarkAsDelete помечает URL на удаление.
func (s *Shortener) MarkAsDelete(ctx context.Context, in *proto.MarkDeletedURLs) (*emptypb.Empty, error) {
	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		s.log.Error("error getting user from context")
		return nil, status.Error(codes.Internal, "")
	}
	if !user.Service.IsAuthenticated {
		return nil, status.Error(codes.Unauthenticated, "")
	}

	go func() {
		err := s.store.AddDeleteTask(in.GetShortUrls(), user.ID)
		if err != nil {
			s.log.Error("error adding new delete task in storage", zap.Error(err))
		}
	}()

	return nil, status.Error(codes.OK, "")
}

// GetServiceStats возвращает статистику сервиса.
func (s *Shortener) GetServiceStats(ctx context.Context, _ *emptypb.Empty) (*proto.GetServiceStatsResponse, error) {
	var (
		res         proto.GetServiceStatsResponse
		usrIPHeader string
		usrIP       net.IP
		trustedNet  *net.IPNet
		err         error
		usersCount  int
		urlsCount   int
	)

	if usrIPHeader, _ = ctx.Value("X-Real-IP").(string); usrIPHeader == "" {
		s.log.Error("empty X-Real-IP header in stats request")
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	usrIP = net.ParseIP(usrIPHeader)
	if usrIP == nil {
		s.log.Error("error parsing IP from header", zap.String("IP header", usrIPHeader))
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	_, trustedNet, err = net.ParseCIDR(s.cfg.TrustedSubNet)
	if err != nil {
		s.log.Error("error parsing CIDR from config", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	if !trustedNet.Contains(usrIP) {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	usersCount, err = s.store.GetUsersCount(ctx)
	if err != nil {
		s.log.Error("error getting users count", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}
	urlsCount, err = s.store.GetURLsCount(ctx)
	if err != nil {
		s.log.Error("error getting urls count", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	res.Users = int32(usersCount)
	res.Urls = int32(urlsCount)

	return &res, nil
}

// GetUserURLs возвращает URL пользователя.
func (s *Shortener) GetUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.GetUserURLsResponse, error) {
	var (
		res  proto.GetUserURLsResponse
		user *models.User
		err  error
	)

	user, ok := ctx.Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		s.log.Error("error getting user from context")
		return nil, status.Error(codes.Internal, "")
	}
	if !user.Service.IsAuthenticated {
		user, err = service.AddNewUser(ctx, s.store, s.cfg)
		if err != nil {
			return nil, status.Error(codes.Internal, "")
		}
	}

	err = grpc.SendHeader(ctx, metadata.New(map[string]string{
		"authorization": user.Service.Token,
	}))
	if err != nil {
		s.log.Error("error adding auth metadata in grpc response", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "error returning token")
	}

	urls, err := s.store.GetURLsByUserID(ctx, user.ID)
	if err != nil {
		s.log.Error("error getting user URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "error collecting URLs")
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.OutOfRange, "no content")
	}

	for _, url := range urls {
		res.UserUrls = append(res.UserUrls, &proto.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &res, nil
}

// Ping пингует базу данных.
func (s *Shortener) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.store.Ping(ctx); err != nil {
		s.log.Error("database is unavailable from ping method", zap.Error(err))
		return nil, status.Error(codes.Internal, "can not ping database")
	}

	return nil, status.Error(codes.OK, "")
}
