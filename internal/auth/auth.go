package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString(userID int, secretKey string, tokenLifeTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifeTime)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error creating signed JWT %w", err)
	}

	return tokenString, nil
}

func GetUserID(tokenString string, secretKey string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected singing method %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

	if err != nil {
		return -1, fmt.Errorf("error parsing tokenString %w", err)
	}

	if !token.Valid {
		return -1, errors.New("token invalid")
	}

	return claims.UserID, nil
}

func GenerateAuthKey() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error getting rand.Read(): %w", err)
	}

	return hex.EncodeToString(b), nil
}
