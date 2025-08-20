package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	authDomain "github.com/Haerd-Limited/dating-api/internal/auth/domain"
)

type RefreshToken struct {
	ID        string // uuid
	UserID    string
	Token     string // secure random string
	ExpiresAt time.Time
}

func GenerateAccessToken(userID string, jwtSecret []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken(userID string) *authDomain.RefreshToken {
	return &authDomain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
}

func ParseAccessToken(tokenStr string, jwtSecret []byte) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if token == nil {
		return "", errors.New("token is nil")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}

	return "", err
}
