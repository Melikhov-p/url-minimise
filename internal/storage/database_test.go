package storage

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseStorage_GetURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() {
		_ = db.Close()
	}()

	storage := DatabaseStorage{
		DB: db,
	}

	testCases := []struct {
		name         string
		shortURL     string
		mockBehavior func(sqlmock.Sqlmock, string)
		wantFound    bool
		wantErr      bool
	}{
		{
			name:     "NotFound",
			shortURL: "notexist.com",
			mockBehavior: func(s sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(
					"SELECT original_url, user_id, uuid, is_deleted  FROM url WHERE short_url = ?",
				).WithArgs(short)
			},
			wantFound: false,
			wantErr:   true,
		},
		{
			name:     "Found",
			shortURL: "exist.com",
			mockBehavior: func(s sqlmock.Sqlmock, short string) {
				rows := sqlmock.NewRows([]string{"original_url", "user_id", "uuid", "is_deleted"}).
					AddRow("full", 1, "12", false)
				mock.ExpectQuery(
					"SELECT original_url, user_id, uuid, is_deleted  FROM url WHERE short_url = ?",
				).WithArgs(short).WillReturnRows(rows)
			},
			wantFound: true,
			wantErr:   true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.shortURL)

			fullURL, err := storage.GetURL(context.Background(), test.shortURL)

			if test.wantFound {
				assert.NotNil(t, fullURL)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, fullURL)
				assert.Error(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
