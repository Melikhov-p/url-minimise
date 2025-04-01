package middlewares

import (
	"context"

	"github.com/Melikhov-p/url-minimise/internal/config"
	"github.com/Melikhov-p/url-minimise/internal/contextkeys"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"github.com/Melikhov-p/url-minimise/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryInterceptor перехватчик запросов gRPC.
type UnaryInterceptor struct {
	log   *zap.Logger
	cfg   *config.Config
	store repository.Storage
}

// NewUnaryInterceptor создает нового перхватчика запросов.
func NewUnaryInterceptor(log *zap.Logger, cfg *config.Config, store repository.Storage) *UnaryInterceptor {
	return &UnaryInterceptor{
		log:   log,
		cfg:   cfg,
		store: store,
	}
}

// UnaryAuthInterceptor - interceptor для авторизации в unary RPC запросах.
func (ui *UnaryInterceptor) UnaryAuthInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	// Извлекаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil) // Инициализируем пустые метаданные, если их нет
	}

	var accessToken string
	if authHeader, exists := md["authorization"]; exists && len(authHeader) > 0 {
		accessToken = authHeader[0]
	}

	// Проверяем токен
	user, err := service.AuthUserByToken(accessToken, ui.store, ui.log, ui.cfg)

	if err != nil {
		// Генерируем новый токен и пользователя
		user = repository.NewEmptyUser()
		ui.log.Info("new empty user")
	} else {
		ui.log.Info("Verified user")
	}

	// Передаем userID в контекст
	ctxWithUser := context.WithValue(ctx, contextkeys.ContextUserKey, user)

	// Передаем управление следующему хендлеру
	return handler(ctxWithUser, req)
}
