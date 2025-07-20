package middleware

import (
	"fmt"
	"log/slog"
	"strings"
	"travel-api/internal/config"
	"travel-api/internal/domain"
	"travel-api/internal/interface/presenter"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("Authorization header missing")
			c.JSON(presenter.ConvertToHTTPError(domain.ErrInvalidCredentials))
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			slog.Warn("Invalid Authorization header format", "header", authHeader)
			c.JSON(presenter.ConvertToHTTPError(domain.ErrInvalidCredentials))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		jwtSecret, err := config.JWTSecret()
		if err != nil {
			c.JSON(presenter.ConvertToHTTPError(domain.NewInternalServerError(err)))
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
			c.JSON(presenter.ConvertToHTTPError(domain.ErrInvalidCredentials))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// ユーザーIDをGinのコンテキストに設定
			c.Set("user_id", claims["user_id"])
			c.Next()
		} else {
			slog.Warn("Invalid JWT claims or token not valid", "claims", claims, "valid", token.Valid)
			c.JSON(presenter.ConvertToHTTPError(domain.ErrInvalidCredentials))
			c.Abort()
			return
		}
	}
}
