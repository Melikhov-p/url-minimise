package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type MemoryStorage struct {
	urls       map[string]*models.StorageURL
	users      map[int]*models.User
	lastUserID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls:       map[string]*models.StorageURL{},
		users:      map[int]*models.User{},
		lastUserID: 0,
	}
}

func (s *MemoryStorage) AddURL(ctx context.Context, newURL *models.StorageURL) (string, error) {
	if short, ok := s.checkFull(ctx, newURL.OriginalURL); ok {
		return short, ErrOriginalURLExist
	}
	s.urls[newURL.ShortURL] = newURL
	return newURL.ShortURL, nil
}

func (s *MemoryStorage) AddURLs(_ context.Context, newURLs []*models.StorageURL) error {
	for _, url := range newURLs {
		s.urls[url.ShortURL] = url
	}

	return nil
}

func (s *MemoryStorage) MarkAsDeletedURL(_ context.Context, inCh chan MarkDeleteURL) chan MarkDeleteResult {
	outCh := make(chan MarkDeleteResult)

	go func() {
		defer close(outCh)

		for url := range inCh {
			if !s.CheckShort(context.Background(), url.ShortURL) {
				outCh <- MarkDeleteResult{
					URL: url.ShortURL,
					Res: false,
					Err: ErrNotFound,
				}
			} else {
				s.urls[url.ShortURL].DeletedFlag = true
				outCh <- MarkDeleteResult{
					URL: url.ShortURL,
					Res: true,
					Err: nil,
				}
			}
		}
	}()

	return outCh
}

func (s *MemoryStorage) GetShortURL(_ context.Context, _ *sql.Tx, fullURL string) (string, error) {
	var short string

	for _, url := range s.urls {
		if url.OriginalURL == fullURL {
			short = url.ShortURL
			return short, nil
		}
	}

	return "", fmt.Errorf("can not found short url for original %w", ErrNotFound)
}

func (s *MemoryStorage) GetURL(_ context.Context, shortURL string) (*models.StorageURL, error) {
	searchedElem := s.urls[shortURL]
	if searchedElem != nil {
		return searchedElem, nil
	}
	return nil, fmt.Errorf("can not found original url for short %w", ErrNotFound)
}

func (s *MemoryStorage) CheckShort(_ context.Context, short string) bool { return s.urls[short] != nil }

// Если оригинальный URL есть в базе - true.
func (s *MemoryStorage) checkFull(_ context.Context, fullURL string) (string, bool) {
	for _, url := range s.urls {
		if fullURL == url.OriginalURL {
			return url.ShortURL, true
		}
	}

	return "", false
}

func (s *MemoryStorage) Ping(_ context.Context) error { return nil }

func (s *MemoryStorage) AddUser(_ context.Context) (*models.User, error) {
	s.lastUserID++
	s.users[s.lastUserID] = &models.User{
		ID:   s.lastUserID,
		URLs: make([]*models.StorageURL, 0),
		Service: &models.UserService{
			IsAuthenticated: false,
			Token:           "",
		},
	}

	return s.users[s.lastUserID], nil
}

func (s *MemoryStorage) GetURLsByUserID(_ context.Context, userID int) ([]*models.StorageURL, error) {
	return s.users[userID].URLs, nil
}
