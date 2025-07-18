# `injector` ディレクトリ

`injector` ディレクトリは、アプリケーションの依存性注入（DI）を管理します。
ここでは、各レイヤーのコンポーネント（リポジトリ、ユースケース、ハンドラーなど）を組み立て、それらの依存関係を解決します。

## `injector.go`

このファイルは、特定のハンドラーやサービスに必要な依存関係を構築し、提供する役割を担います。

### `NewTripHandler` 関数

`NewTripHandler` 関数は、`Trip`関連のハンドラーを初期化し、その依存関係を注入します。

-   **`db postgres.DBTX`**: データベーストランザクションインターフェースを受け取ります。これにより、リポジトリがデータベースと連携できるようになります。
-   **`tripRepository`**: `postgres.NewTripPostgresRepository` を使用して、`Trip`エンティティの永続化を担うリポジトリの実装を生成します。
-   **`clock`**: `domain.SystemClock` のインスタンスを生成し、現在時刻を提供するサービスとして注入します。
-   **`uuidGenerator`**: `domain.DefaultUUIDGenerator` のインスタンスを生成し、UUIDを生成するサービスとして注入します。
-   **`tripUsecase`**: `usecase.NewTripInteractor` を使用して、`Trip`関連のビジネスロジックをカプセル化するユースケースを生成し、上記で作成したリポジトリ、クロック、UUIDジェネレーターを注入します。
-   **`handler.NewTripHandler`**: 最終的に、構築された`tripUsecase`を`Trip`ハンドラーに注入し、初期化されたハンドラーを返します。

この仕組みにより、各コンポーネントは具体的な実装に依存せず、インターフェースを通じて連携するため、テスト容易性や保守性が向上します。

```go
package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/infrastructure/postgres"
	"travel-api/internal/interface/handler"
	"travel-api/internal/usecase"
)

func NewTripHandler(db postgres.DBTX) *handler.TripHandler {
	tripRepository := postgres.NewTripPostgresRepository(db)
	clock := &domain.SystemClock{}
	uuidGenerator := &domain.DefaultUUIDGenerator{}
	tripUsecase := usecase.NewTripInteractor(tripRepository, clock, uuidGenerator)
	return handler.NewTripHandler(tripUsecase)
}
```
