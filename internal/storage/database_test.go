package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Melikhov-p/url-minimise/internal/models"
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

func TestDatabaseStorage_AddURL(t *testing.T) {
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
		newURL       *models.StorageURL
		shortURL     string
		mockBehavior func(sqlmock.Sqlmock, *models.StorageURL)
		wantErr      bool
		wantErrIs    error
	}{
		{
			name: "SuccessAdd",
			newURL: &models.StorageURL{
				OriginalURL: "original",
				ShortURL:    "short",
				UserID:      1,
				UUID:        "1",
				DeletedFlag: false,
			},
			shortURL: "short",
			mockBehavior: func(s sqlmock.Sqlmock, new *models.StorageURL) {
				mock.ExpectBegin()

				prepared := mock.ExpectPrepare(
					"SELECT short_url  FROM url WHERE original_url = ?",
				)

				prepared.ExpectQuery().WithArgs(new.OriginalURL).WillReturnError(sql.ErrNoRows)

				preparedInsert := mock.ExpectPrepare("INSERT INTO URL")
				preparedInsert.ExpectExec().WithArgs(new.ShortURL, new.OriginalURL, new.UserID, new.DeletedFlag).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantErr:   false,
			wantErrIs: nil,
		},
		{
			name: "URL Exist",
			newURL: &models.StorageURL{
				OriginalURL: "original",
				ShortURL:    "short",
				UserID:      1,
				UUID:        "1",
				DeletedFlag: false,
			},
			shortURL: "short",
			mockBehavior: func(s sqlmock.Sqlmock, new *models.StorageURL) {
				mock.ExpectBegin()

				prepared := mock.ExpectPrepare(
					"SELECT short_url  FROM url WHERE original_url = ?",
				)

				rows := mock.NewRows([]string{"short_url"}).AddRow("short")
				prepared.ExpectQuery().WithArgs(new.OriginalURL).WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr:   true,
			wantErrIs: ErrOriginalURLExist,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.newURL)

			shortURL, err := storage.AddURL(context.Background(), test.newURL)
			if test.wantErr {
				assert.ErrorIs(t, err, test.wantErrIs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.shortURL, shortURL)
			}
		})
	}
}

func TestDatabaseStorage_AddURLs(t *testing.T) {
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
		newURLs      []*models.StorageURL
		shortURL     string
		mockBehavior func(sqlmock.Sqlmock, []*models.StorageURL)
		wantErr      bool
		wantErrIs    error
	}{
		{
			name: "SuccessAdd",
			newURLs: []*models.StorageURL{
				{
					OriginalURL: "original",
					ShortURL:    "short",
					UserID:      1,
					UUID:        "1",
					DeletedFlag: false,
				},
			},
			shortURL: "short",
			mockBehavior: func(s sqlmock.Sqlmock, news []*models.StorageURL) {
				for _, newURL := range news {
					mock.ExpectBegin()
					preparedInsert := mock.ExpectPrepare("INSERT INTO url")
					preparedInsert.ExpectExec().WithArgs(
						newURL.ShortURL,
						newURL.OriginalURL,
						newURL.UserID,
						newURL.DeletedFlag,
					).
						WillReturnResult(sqlmock.NewResult(1, 1))

					mock.ExpectCommit()
				}
			},
			wantErr:   false,
			wantErrIs: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.newURLs)

			err = storage.AddURLs(context.Background(), test.newURLs)
			if test.wantErr {
				assert.ErrorIs(t, err, test.wantErrIs)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_AddUser(t *testing.T) {
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
		mockBehavior func(sqlmock.Sqlmock)
		wantErr      bool
	}{
		{
			name: "Success",
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectBegin()

				rows := mock.NewRows([]string{"id"}).AddRow(1)

				mock.ExpectQuery(`INSERT INTO "user" DEFAULT VALUES RETURNING id`).
					WillReturnRows(rows)

				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Fail",
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectBegin()

				rows := mock.NewRows([]string{"id"}).AddRow("asdasq")

				mock.ExpectQuery(`INSERT INTO "user" DEFAULT VALUES RETURNING id`).
					WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock)

			user, err := storage.AddUser(context.Background())
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_GetSecretKey(t *testing.T) {
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
		mockBehavior func(sqlmock.Sqlmock)
		secretKey    string
		wantErr      bool
	}{
		{
			name:      "Key exist",
			secretKey: "secret",
			wantErr:   false,
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectBegin()

				rows := mock.NewRows([]string{"key"}).AddRow("secret")
				mock.ExpectQuery(`SELECT key FROM secret_key`).WillReturnRows(rows)

				mock.ExpectCommit()
			},
		},
		{
			name:      "Key not exist",
			secretKey: "secret",
			wantErr:   false,
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT key FROM secret_key`).WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(`INSERT INTO secret_key`).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock)

			secret, err := storage.GetSecretKey(context.Background())

			assert.NoError(t, err)
			assert.NotEmpty(t, secret)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
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
		mockBehavior func(sqlmock.Sqlmock)
		wantErr      bool
	}{
		{
			name:    "Unavailable",
			wantErr: true,
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("database is not available"))
			},
		},
		{
			name:    "Available",
			wantErr: false,
			mockBehavior: func(s sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
		},
	}

	for _, test := range testCases {
		test.mockBehavior(mock)

		err = storage.Ping(context.Background())

		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	}
}

func TestDatabaseStorage_AddDeleteTask(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
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
		mockBehavior func(sqlmock.Sqlmock, []string)
		shorts       []string
		userID       int
		wantErr      bool
	}{
		{
			name:    "Success",
			userID:  1,
			wantErr: false,
			shorts:  []string{"short1", "short2", "short3"},
			mockBehavior: func(s sqlmock.Sqlmock, shorts []string) {
				mock.ExpectBegin()
				prepared := mock.ExpectPrepare(`INSERT INTO delete_task`)

				for _, short := range shorts {
					prepared.ExpectExec().WithArgs(short, 1, models.Registered).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}

				mock.ExpectCommit()
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.shorts)

			err = storage.AddDeleteTask(test.shorts, test.userID)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestDatabaseStorage_CheckShort(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
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
		short        string
		mockBehavior func(sqlmock.Sqlmock, string)
		exist        bool
	}{
		{
			name:  "URL exist",
			short: "short",
			exist: true,
			mockBehavior: func(s sqlmock.Sqlmock, short string) {

				rows := mock.NewRows([]string{"original_url", "user_id", "uuid", "is_deleted"}).
					AddRow("original", 1, "1", false)
				mock.ExpectQuery(`SELECT original_url, user_id, uuid, is_deleted  FROM url WHERE short_url = ?`).
					WithArgs(short).WillReturnRows(rows)

			},
		},
		{
			name:  "URL not exist",
			short: "short",
			exist: false,
			mockBehavior: func(s sqlmock.Sqlmock, short string) {
				mock.ExpectQuery(`SELECT original_url, user_id, uuid, is_deleted  FROM url WHERE short_url = ?`).
					WithArgs(short).WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.short)

			exist := storage.CheckShort(context.Background(), test.short)
			assert.Equal(t, test.exist, exist)
		})
	}
}

func TestDatabaseStorage_GetShortURL(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
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
		name          string
		expectedShort string
		wantErr       bool
		wantErrIs     error
		fullUrl       string
		mockBehavior  func(sqlmock.Sqlmock, string)
	}{
		{
			name:          "Success",
			expectedShort: "short",
			wantErr:       false,
			wantErrIs:     nil,
			fullUrl:       "full",
			mockBehavior: func(s sqlmock.Sqlmock, full string) {
				prepared := mock.ExpectPrepare(`SELECT short_url FROM url WHERE original_url = ?`)

				rows := mock.NewRows([]string{"short_url"}).AddRow("short")
				prepared.ExpectQuery().WithArgs(full).WillReturnRows(rows)

			},
		},
		{
			name:          "Not Found",
			expectedShort: "short",
			wantErr:       true,
			wantErrIs:     ErrNotFound,
			fullUrl:       "full",
			mockBehavior: func(s sqlmock.Sqlmock, full string) {
				prepared := mock.ExpectPrepare(`SELECT short_url FROM url WHERE original_url = ?`)

				prepared.ExpectQuery().WithArgs(full).WillReturnError(sql.ErrNoRows)

			},
		},
	}
	mock.ExpectBegin()
	tx, err := storage.DB.Begin()
	assert.NoError(t, err)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			mock.ExpectRollback()
		} else {
			_ = tx.Commit()
			mock.ExpectCommit()
		}
	}()

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.fullUrl)
			ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
			defer cancel()

			short, err := storage.GetShortURL(ctx, tx, test.fullUrl)
			if test.wantErr {
				assert.ErrorIs(t, err, test.wantErrIs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedShort, short)
			}
		})
	}
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDatabaseStorage_GetDeleteTasksWStatus(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
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
		wantErr      bool
		status       models.DelTaskStatus
		mockBehavior func(sqlmock.Sqlmock, models.DelTaskStatus)
	}{
		{
			name:    "Success",
			wantErr: false,
			status:  models.DelTaskStatus("Registered"),
			mockBehavior: func(s sqlmock.Sqlmock, status models.DelTaskStatus) {
				rows := sqlmock.NewRows([]string{"short_url", "user_id"}).AddRow("short", 1)

				mock.ExpectQuery(`SELECT short_url, user_id FROM delete_task WHERE status=?`).
					WithArgs(status).WillReturnRows(rows)
			},
		},
		{
			name:    "Fail",
			wantErr: true,
			status:  models.DelTaskStatus("Registered"),
			mockBehavior: func(s sqlmock.Sqlmock, status models.DelTaskStatus) {

				mock.ExpectQuery(`SELECT short_url, user_id FROM delete_task WHERE status=?`).
					WithArgs(status).WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(mock, test.status)
			tasks, err := storage.GetDeleteTasksWStatus(context.Background(), test.status)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tasks)
			}
		})
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	}
}
