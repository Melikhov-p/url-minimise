package repository

import (
	"github.com/Melikhov-p/url-minimise/internal/models"
)

func NewEmptyUser() *models.User {
	return &models.User{
		ID:   -1,
		URLs: make([]*models.StorageURL, 0),
		Service: &models.UserService{
			IsAuthenticated: false,
			Token:           "",
		},
	}
}
