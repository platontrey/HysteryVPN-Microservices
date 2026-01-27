package utils

import (
	"fmt"
	"time"

	"hysteria2_microservices/api-service/internal/services/interfaces"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateJWT(userID uuid.UUID, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &interfaces.Claims{
		UserID:   userID.String(),
		Username: "", // Will be filled by service
		Role:     "", // Will be filled by service
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(expiry).Unix(),
		"iss":      "hysteria2-api",
		"sub":      userID.String(),
	})

	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString, secret string) (*interfaces.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)

	return &interfaces.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}
