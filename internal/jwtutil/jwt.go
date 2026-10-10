package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT claims for a user.
type Claims struct {
	UserID uint
	jwt.RegisteredClaims
}

// TokenManager handles JWT token generation and validation.
type TokenManager struct {
	secret     string
	expiration time.Duration
}

// NewTokenManager creates a new token manager.
func NewTokenManager(secret string, expiration time.Duration) *TokenManager {
	return &TokenManager{
		secret:     secret,
		expiration: expiration,
	}
}

// GenerateToken generates a new JWT token for a user.
func (tm *TokenManager) GenerateToken(userID uint) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tm.secret))
}

// ValidateToken validates and parses a JWT token.
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tm.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
