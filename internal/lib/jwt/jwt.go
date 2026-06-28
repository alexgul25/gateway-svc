package jwtpt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSigningMethod = errors.New("unexpected signing method")
	ErrParseToken    = errors.New("failed to parse token")
)

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func ParseToken(tokenStr string, secret []byte) (*TokenClaims, error) {
	var claims TokenClaims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrSigningMethod
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseToken, err)
	}

	return &claims, nil
}
