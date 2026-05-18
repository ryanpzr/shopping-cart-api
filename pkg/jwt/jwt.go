package jwtutil

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
}

func Generate(userID int, role, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Role:   role,
		Email:  email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", apperrors.NewInternalServer("failed to generate token")
	}
	return signed, nil
}

func Parse(tokenStr string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.NewUnauthorized("invalid token signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperrors.NewUnauthorized("invalid or expired token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, apperrors.NewUnauthorized("invalid token claims")
	}
	return claims, nil
}
