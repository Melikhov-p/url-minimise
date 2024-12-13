package service

import (
	"context"
	"fmt"

	"github.com/Melikhov-p/url-minimise/internal/auth"
	"github.com/Melikhov-p/url-minimise/internal/models"
	"github.com/Melikhov-p/url-minimise/internal/repository"
	"go.uber.org/zap"
)

func AuthUserByToken(tokenString string, s repository.Storage, logger *zap.Logger) (*models.User, error) {
	emptyUser := repository.NewEmptyUser()
	userID, err := auth.GetUserID(tokenString)
	logger.Debug("auth by token", zap.String("TOKEN", tokenString), zap.Int("ID", userID))
	if err != nil {
		return emptyUser, fmt.Errorf("error getting user ID from token %w", err)
	}

	var urls []*models.StorageURL
	urls, err = s.GetURLsByUserID(context.Background(), userID)
	if err != nil {
		logger.Error("error getting urls by user id", zap.Error(err))
	}

	emptyUser.URLs = urls
	emptyUser.ID = userID
	emptyUser.Service.IsAuthenticated = true
	emptyUser.Service.Token = tokenString
	return emptyUser, nil
}

func AddNewUser(ctx context.Context, s repository.Storage) (*models.User, error) {
	user, err := s.AddUser(ctx)
	if err != nil {
		return repository.NewEmptyUser(), fmt.Errorf("error creating new user in storage %w", err)
	}

	user.Service.Token, err = BuildUserToken(user.ID)
	if err != nil {
		return user, fmt.Errorf("error creating token for new user %w", err)
	}

	user.Service.IsAuthenticated = true
	return user, nil
}

func BuildUserToken(userID int) (string, error) {
	// BuildUserToken return token string for userID int
	token, err := auth.BuildJWTString(userID)
	if err != nil {
		return "", fmt.Errorf("error creating token for user %w", err)
	}

	return token, nil
}
