package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const SECRETKEY = "supersecretkey"
const tokenLifeTime = 24 * time.Hour

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifeTime)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SECRETKEY))
	if err != nil {
		return "", fmt.Errorf("error creating signed JWT %w", err)
	}

	return tokenString, nil
}

func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected singing method %v", t.Header["alg"])
			}
			return []byte(SECRETKEY), nil
		})

	if err != nil {
		return -1, fmt.Errorf("error parsing tokenString %w", err)
	}

	if !token.Valid {
		return -1, errors.New("token invalid")
	}

	return claims.UserID, nil
}
