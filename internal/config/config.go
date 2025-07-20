package config

import (
	"fmt"
	"os"
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
