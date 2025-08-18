package auth

import (
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

func GenerateRefreshToken(userID string) (*authDomain.RefreshToken, error) {
	id := uuid.New().String()
	token := uuid.New().String()
	expires := time.Now().Add(7 * 24 * time.Hour)

	rt := &authDomain.RefreshToken{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expires,
	}

	return rt, nil
}

func ParseAccessToken(tokenStr string, jwtSecret []byte) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}

	return "", err
}
