package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hata0/travel-api/internal/adapter/presenter"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("Authorization header missing")
			c.JSON(presenter.ConvertToHTTPError(
				apperr.NewInvalidCredentialsError("authorization header is required"),
			))
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			slog.Warn("Invalid Authorization header format", "header", authHeader)
			c.JSON(presenter.ConvertToHTTPError(
				apperr.NewInvalidCredentialsError("invalid authorization header format"),
			))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 署名方法の検証
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			slog.Warn("JWT token validation failed", "error", err)
			c.JSON(presenter.ConvertToHTTPError(
				apperr.NewInvalidCredentialsError("jwt token validation failed"),
			))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// ユーザーIDをGinのコンテキストに設定
			c.Set("user_id", claims["user_id"])
			c.Next()
		} else {
			slog.Warn("Invalid JWT claims or token not valid", "claims", claims, "valid", token.Valid)
			c.JSON(presenter.ConvertToHTTPError(
				apperr.NewInvalidCredentialsError("invalid jwt claims or token not valid"),
			))
			c.Abort()
			return
		}
	}
}
