# 認証処理の概要

このドキュメントでは、`travel-api`における認証・認可機能の全体像と、各コンポーネントの役割について説明します。

## 1. ユーザー登録 (Register)

ユーザーが新しいアカウントを作成するプロセスです。

-   **インターフェース層 (`internal/interface/handler/auth.go`)**:
    -   `/register` エンドポイントでリクエストを受け付けます。
    -   `internal/interface/validator/auth.go` で定義された `RegisterJSONBody` を使用して、リクエストボディのバリデーションを行います。
    -   `AuthUsecase` の `Register` メソッドを呼び出します。
    -   成功した場合、`http.StatusCreated` (201) とユーザーIDを返します。
    -   バリデーションエラー、ユーザー名またはメールアドレスの重複、内部エラーが発生した場合は、適切なエラーレスポンスを返します。

-   **ユースケース層 (`internal/usecase/auth.go`)**:
    -   `AuthInteractor` の `Register` メソッドがビジネスロジックを処理します。
    -   `UserRepository` を使用して、ユーザー名とメールアドレスの重複をチェックします。
    -   `golang.org/x/crypto/bcrypt` を使用してパスワードをハッシュ化します。
    -   `UUIDGenerator` を使用して新しいユーザーIDを生成します。
    -   `Clock` を使用して `CreatedAt` と `UpdatedAt` のタイムスタンプを設定します。
    -   `UserRepository` を使用して新しいユーザーをデータベースに保存します。
    -   成功した場合、`output.RegisterOutput` (ユーザーIDを含む) を返します。

-   **ドメイン層 (`internal/domain/user.go`)**:
    -   `User` エンティティがユーザー情報を表現します。
    -   `UserID` はユーザーIDの値オブジェクトです。
    -   `UserRepository` インターフェースがユーザーの永続化操作を抽象化します。

-   **インフラ層 (`internal/infrastructure/postgres/user_postgres.go`)**:
    -   `UserPostgresRepository` が `UserRepository` インターフェースを実装します。
    -   `sqlc` が生成したクエリ (`internal/infrastructure/postgres/sql/queries/users.sql`) を使用してデータベース操作を行います。

## 2. ユーザーログイン (Login)

ユーザーが認証情報を提示してシステムにログインするプロセスです。

-   **インターフェース層 (`internal/interface/handler/auth.go`)**:
    -   `/login` エンドポイントでリクエストを受け付けます。
    -   `internal/interface/validator/auth.go` で定義された `LoginJSONBody` を使用して、リクエストボディのバリデーションを行います。
    -   `AuthUsecase` の `Login` メソッドを呼び出します。
    -   成功した場合、`http.StatusOK` (200) とアクセストークン、リフレッシュトークンを返します。
    -   バリデーションエラー、認証情報の不一致、内部エラーが発生した場合は、適切なエラーレスポンスを返します。

-   **ユースケース層 (`internal/usecase/auth.go`)**:
    -   `AuthInteractor` の `Login` メソッドがビジネスロジックを処理します。
    -   `UserRepository` を使用してユーザーをメールアドレスで検索します。
    -   `bcrypt` を使用して提示されたパスワードと保存されているハッシュ化されたパスワードを比較し、検証します。
    -   `github.com/golang-jwt/jwt/v5` を使用してアクセストークンを生成します。
    -   `UUIDGenerator` を使用してリフレッシュトークンを生成します。
    -   `RefreshTokenRepository` を使用してリフレッシュトークンをデータベースに保存します。
    -   成功した場合、`output.LoginOutput` (アクセストークンとリフレッシュトークンを含む) を返します。

-   **ドメイン層 (`internal/domain/refresh_token.go`)**:
    -   `RefreshToken` エンティティがリフレッシュトークン情報を表現します。
    -   `RefreshTokenRepository` インターフェースがリフレッシュトークンの永続化操作を抽象化します。

-   **インフラ層 (`internal/infrastructure/postgres/refresh_token_postgres.go`)**:
    -   `RefreshTokenPostgresRepository` が `RefreshTokenRepository` インターフェースを実装します。
    -   `sqlc` が生成したクエリ (`internal/infrastructure/postgres/sql/queries/refresh_tokens.sql`) を使用してデータベース操作を行います。

## 3. トークンリフレッシュ (Token Refresh)

アクセストークンの有効期限が切れた際に、リフレッシュトークンを使用して新しいアクセストークンとリフレッシュトークンを取得するプロセスです。

-   **ユースケース層 (`internal/usecase/auth.go`)**:
    -   `AuthInteractor` の `VerifyRefreshToken` メソッドがリフレッシュロジックを処理します。
    -   `RefreshTokenRepository` を使用してリフレッシュトークンをデータベースから検索し、有効期限をチェックします。
    -   有効なリフレッシュトークンであれば、古いトークンを削除し、新しいアクセストークンとリフレッシュトークンを生成して返します。
    -   `UserRepository` を使用してユーザー情報を取得します。

## 4. トークン失効 (Token Revocation)

リフレッシュトークンを強制的に無効化するプロセスです（例: ログアウト時）。

-   **ユースケース層 (`internal/usecase/auth.go`)**:
    -   `AuthInteractor` の `RevokeRefreshToken` メソッドがトークン失効ロジックを処理します。
    -   `RefreshTokenRepository` を使用してリフレッシュトークンをデータベースから削除します。

## 5. 認証ミドルウェア (Authentication Middleware)

保護されたAPIエンドポイントへのアクセスを制御します。

-   **インターフェース層 (`internal/interface/middleware/auth.go`)**:
    -   `AuthMiddleware` はGinのミドルウェアとして機能します。
    -   HTTPリクエストの `Authorization` ヘッダーからJWTアクセストークンを抽出します。
    -   `github.com/golang-jwt/jwt/v5` を使用してトークンの署名と有効性を検証します。
    -   トークンが有効であれば、ユーザーIDをGinのコンテキストに設定し、次のハンドラに処理を渡します。
    -   トークンが無効または欠落している場合は、`http.StatusUnauthorized` (401) または `http.StatusBadRequest` (400) のエラーレスポンスを返します。
    -   `log/slog` を使用して、認証失敗の各シナリオで警告またはエラーログを出力します。

## 6. 依存性注入 (Dependency Injection)

-   **`internal/injector/injector.go`**:
    -   `NewAuthHandler` 関数が `AuthHandler` と `AuthInteractor` の依存関係を解決し、構築します。
    -   `AuthInteractor` には `UserRepository`, `RefreshTokenRepository`, `Clock`, `UUIDGenerator` が注入されます。

## 7. ルーティング (`cmd/server/main.go`)

-   `main.go` で `AuthHandler` を登録し、認証が必要なAPIグループに `AuthMiddleware` を適用します。
