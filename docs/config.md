# `config` ディレクトリ

`config` ディレクトリは、アプリケーションの設定管理を担います。
主に環境変数からの設定値の読み込みや、設定値のバリデーションなどを行います。

## `config.go`

このファイルは、データベース接続文字列（DSN）の取得機能を提供します。

### `DSN()` 関数

`DSN()` 関数は、データベース接続文字列を環境変数から取得します。

-   `.env` ファイルからの読み込み: `godotenv.Load()` を使用して、プロジェクトルートにある `.env` ファイルから環境変数を読み込みます。これにより、開発環境での設定が容易になります。
-   環境変数の優先: `.env` ファイルが存在しない場合や、環境変数が直接設定されている場合は、そちらが優先されます。
-   デフォルト値: `DB_USER`, `DB_PASS`, `DB_HOST`, `DB_PORT`, `DB_NAME` のいずれかの環境変数が設定されていない場合、開発用のデフォルトDSN (`postgres://dev_user:dev_pass@localhost:5432/dev_db`) を返します。
-   DSNの構築: 全ての環境変数が設定されている場合、それらの値を使用してPostgreSQLの接続文字列を構築し、返します。

この関数により、アプリケーションは環境に依存しない形でデータベース接続情報を取得できます。

```go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"travel-api/internal/domain"
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
```
