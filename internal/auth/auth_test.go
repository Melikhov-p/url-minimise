package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildJWTString(t *testing.T) {
	userID := 123
	secretKey := "mysecretkey"
	tokenLifeTime := time.Hour

	tokenString, err := BuildJWTString(userID, secretKey, tokenLifeTime)

	assert.NoError(t, err, "Expected no error when building JWT string")
	assert.NotEmpty(t, tokenString, "Expected non-empty token string")

	// Проверяем, что токен можно распарсить и получить из него userID
	parsedUserID, err := GetUserID(tokenString, secretKey)
	assert.NoError(t, err, "Expected no error when parsing JWT string")
	assert.Equal(t, userID, parsedUserID, "Expected user ID to match")
}

func TestGetUserID(t *testing.T) {
	userID := 456
	secretKey := "mysecretkey"
	tokenLifeTime := time.Hour

	tokenString, err := BuildJWTString(userID, secretKey, tokenLifeTime)
	assert.NoError(t, err, "Expected no error when building JWT string")

	parsedUserID, err := GetUserID(tokenString, secretKey)
	assert.NoError(t, err, "Expected no error when parsing JWT string")
	assert.Equal(t, userID, parsedUserID, "Expected user ID to match")

	// Проверка с неверным ключом
	_, err = GetUserID(tokenString, "wrongkey")
	assert.Error(t, err, "Expected error when parsing JWT string with wrong key")
}

func TestGenerateAuthKey(t *testing.T) {
	key, err := GenerateAuthKey()

	assert.NoError(t, err, "Expected no error when generating auth key")
	assert.Len(t, key, 32, "Expected auth key to be 32 characters long")
}
