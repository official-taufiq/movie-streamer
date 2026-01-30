package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	Role string
	jwt.RegisteredClaims
}

func MakeJwt(userID, secret, role string, expiry time.Duration) (string, error) {

	claims := &AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "movie-streamer",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(token), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header included")
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("malformed authorization header")
	}

	return parts[1], nil
}

func ValidateJwt(jwtToken, jwtSecret string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
