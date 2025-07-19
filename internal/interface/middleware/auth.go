package middleware

import (
	"fmt"
	"log/slog"
	"strings"
	"travel-api/internal/config"
	"travel-api/internal/domain"
	"travel-api/internal/interface/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("Authorization header missing")
			response.NewError(domain.ErrInvalidCredentials).JSON(c)
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			slog.Warn("Invalid Authorization header format", "header", authHeader)
			response.NewError(domain.ErrInvalidCredentials).JSON(c)
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		jwtSecret, err := config.JWTSecret()
		if err != nil {
			slog.Error("Failed to get JWT secret", "error", err)
			response.NewError(domain.NewInternalServerError(err)).JSON(c)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 署名方法の検証
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			slog.Warn("JWT token validation failed", "error", err)
			response.NewError(domain.ErrInvalidCredentials).JSON(c)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// ユーザーIDをGinのコンテキストに設定
			c.Set("user_id", claims["user_id"])
			c.Next()
		} else {
			slog.Warn("Invalid JWT claims or token not valid", "claims", claims, "valid", token.Valid)
			response.NewError(domain.ErrInvalidCredentials).JSON(c)
			c.Abort()
			return
		}
	}
}
