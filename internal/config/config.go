package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	"travel-api/internal/domain"

	"github.com/joho/godotenv"
)

// DSN は、データベース接続文字列 (DSN) を環境変数から取得します。
// 環境変数が設定されていない場合は、デフォルトのDSNを返します。
func DSN() (string, error) {
	// .envファイルから環境変数を読み込む
	// エラーが発生しても、環境変数が直接設定されていれば問題ないため、エラーは無視します
	_ = godotenv.Load()

	// 環境変数からデータベース接続情報を取得
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// いずれかの環境変数が設定されていない場合は、デフォルトのDSNを返す
	if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbName == "" {
		return "postgres://dev_user:dev_pass@localhost:5432/dev_db", nil
	}

	// データベース接続文字列を構築
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName), nil
}

// JWTSecret は、JWTの秘密鍵を環境変数から取得します。
// 環境変数が設定されていない場合は、エラーを返します。
func JWTSecret() (string, error) {
	_ = godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", domain.ErrConfiguration
	}
	return jwtSecret, nil
}

// AccessTokenExpiration は、アクセストークンの有効期限を環境変数から取得します。
// 環境変数が設定されていない場合は、デフォルトの24時間を返します。
func AccessTokenExpiration() time.Duration {
	_ = godotenv.Load()
	expStr := os.Getenv("ACCESS_TOKEN_EXPIRATION_HOURS")
	if expStr == "" {
		return 24 * time.Hour // デフォルト値
	}
	hours, err := strconv.Atoi(expStr)
	if err != nil {
		slog.Warn("Invalid ACCESS_TOKEN_EXPIRATION_HOURS, using default", "value", expStr)
		return 24 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

// RefreshTokenExpiration は、リフレッシュトークンの有効期限を環境変数から取得します。
// 環境変数が設定されていない場合は、デフォルトの7日間を返します。
func RefreshTokenExpiration() time.Duration {
	_ = godotenv.Load()
	expStr := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS")
	if expStr == "" {
		return 7 * 24 * time.Hour // デフォルト値
	}
	days, err := strconv.Atoi(expStr)
	if err != nil {
		slog.Warn("Invalid REFRESH_TOKEN_EXPIRATION_DAYS, using default", "value", expStr)
		return 7 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}
