package jwt

import (
	"errors"
	"fmt"
	"time"
	"url_profile/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID   int    `json:"uid"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func NewToken(user *models.User, duration time.Duration, secret string) (string, error) {
	fmt.Printf("Creating token with duration: %v\n", duration)
	claims := Claims{
		UID:   user.ID,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	bearerToken := fmt.Sprintf("Bearer %s", tokenString)
	return bearerToken, nil
}

func ParseAndVerify(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
