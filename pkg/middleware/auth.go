package middleware

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"net/http"
	"news-feed/internal/cache"
	"news-feed/pkg/config"
	"strings"
	"time"
)

var jwtKey = config.LoadConfig().JWTSecret
var redisClient = cache.GetRedisClient()

type Claims struct {
	Subject string `json:"sub"`
	jwt.StandardClaims
}

func JWTAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Authorization header missing", http.StatusUnauthorized)
					return
				}

				// Extract the token from the header
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")

				// Validate the token
				claims, err := ValidateJWT(tokenString)
				if err != nil {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}

				// Check if the token exists in Redis
				exists, err := redisClient.Exists(context.Background(), tokenString).Result()
				if err != nil || exists == 0 {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}

				// Set user ID in context for use in handlers
				ctx := context.WithValue(r.Context(), "userID", claims.Subject)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the token's signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return jwtKey, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GenerateJWT generates a new JWT for a given user ID.
func GenerateJWT(userName string) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   userName,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token valid for 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := config.LoadConfig().JWTSecret
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
