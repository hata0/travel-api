# `cmd` ディレクトリ

`cmd` ディレクトリは、アプリケーションのエントリーポイントを定義します。
各サブディレクトリは独立した実行可能アプリケーションを表します。

## `cmd/server/main.go`

このファイルは、Web APIサーバーのメインエントリーポイントです。
以下の主要な処理を実行します。

-   **データベース接続の初期化**: `internal/config` パッケージからDSNを取得し、PostgreSQLデータベースへの接続プールを初期化します。
-   **Ginルーターの設定**: Ginフレームワークを使用してHTTPルーターをセットアップします。
-   **APIエンドポイントの登録**: `internal/injector` を使用して依存性を注入し、各ハンドラーのAPIエンドポイントをルーターに登録します。
-   **HTTPサーバーの起動**: 設定されたルーターを使用してHTTPサーバーを起動します。
-   **グレースフルシャットダウン**: SIGINT (Ctrl+C) や SIGTERM などのシグナルを受信した際に、サーバーを安全にシャットダウンするための処理を実装しています。

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"travel-api/internal/config"
	"travel-api/internal/injector"
	"travel-api/internal/interface/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn, err := config.DSN()
	if err != nil {
		slog.Error("Failed to get DSN", "error", err)
		os.Exit(1)
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	router := gin.Default()

	injector.NewAuthHandler(db).RegisterAPI(router)
	injector.NewTripHandler(db).RegisterAPI(router)

	// 認証が必要なAPIグループ
	authRequired := router.Group("/")
	authRequired.Use(middleware.AuthMiddleware())
	{
		// ここに認証が必要なAPIを登録
		// 例: authRequired.GET("/protected", handler.ProtectedHandler)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.Info("Server starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to run server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
```
