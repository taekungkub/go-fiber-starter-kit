package core

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// AccessTokenSecret is the secret key for access tokens
	AccessTokenSecret []byte
	// RefreshTokenSecret is the secret key for refresh tokens
	RefreshTokenSecret []byte
	// AccessTokenExpiry defines access token expiration (default 15 min)
	AccessTokenExpiry time.Duration
	// RefreshTokenExpiry defines refresh token expiration (default 7 days)
	RefreshTokenExpiry time.Duration
)

// Claims represents JWT claims for both access and refresh tokens
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a short-lived access token
func GenerateAccessToken(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessTokenSecret)
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(RefreshTokenSecret)
}

// ValidateAccessToken validates an access token and returns the claims
func ValidateAccessToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, AccessTokenSecret)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, RefreshTokenSecret)
}

func validateToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID, email, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userID, email, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateRefreshToken(userID, email, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
