package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Melikhov-p/url-minimise/internal/models"
)

type DatabaseStorage struct {
	DB *sql.DB
}

const dbTimeout = 15 * time.Second

func (db *DatabaseStorage) AddURL(newURL *models.StorageURL) {
	fmt.Println("add url in db")
}

func (db *DatabaseStorage) GetFullURL(shortURL string) string {
	fmt.Println("get full url")
	return ""
}

func (db *DatabaseStorage) CheckEl(string) bool {
	return false
}

func (db *DatabaseStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("error ping context %w", err)
	}

	return nil
}
