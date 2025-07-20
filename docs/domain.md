# ドメイン層の作成

ドメイン層は、アプリケーションの核となるビジネスロジックとエンティティを定義します。
ここでは、特定の永続化技術やフレームワークに依存しない、純粋なビジネスルールを記述します。

## 1. エンティティの定義

ドメインの核となるエンティティ（例: `Trip`）を定義します。
エンティティは、そのドメインオブジェクトが持つべき属性と、その属性に対する操作をメソッドとして持ちます。

-   **ファイルパス**: `internal/domain/<entity_name>.go` (例: `internal/domain/trip.go`)
-   **内容**:
    -   エンティティを表す構造体を定義します。
    -   エンティティのIDには、`TripID`のような値オブジェクトを定義し、型安全性を高めます。
    -   エンティティの生成には、不変性を保つためのコンストラクタ関数（例: `NewTrip`）を定義します。
    -   エンティティの状態を変更する操作は、値レシーバのメソッドとして定義し、新しいインスタンスを返すことで不変性を維持します（例: `Trip.Update`）。

```go
// internal/domain/trip.go の例
type TripID struct {
	value string
}

func NewTripID(id string) (TripID, error) {
	if !IsValidUUID(id) {
		return TripID{}, ErrInvalidUUID
	}
	return TripID{value: id}, nil
}

func (id TripID) String() string {
	return id.value
}

type Trip struct {
	ID        TripID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTrip は新しいTripエンティティを作成します。
func NewTrip(id TripID, name string, createdAt time.Time, updatedAt time.Time) Trip {
	return Trip{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// Update はTripの名前と更新日時を更新し、新しいTripエンティティを返します。
func (t Trip) Update(name string, updatedAt time.Time) Trip {
	return Trip{
		ID:        t.ID,
		Name:      name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: updatedAt,
	}
}
```

## 2. リポジトリインターフェースの定義

エンティティの永続化操作を抽象化するリポジトリインターフェースを定義します。このインターフェースは、ドメイン層をデータベースなどの具体的な永続化層の実装から分離するための重要な役割を担います。

### リポジトリインターフェース実装チェックリスト

新しいリポジトリインターフェースを定義する際は、以下の項目を確認してください。

-   [ ] **ファイルと場所**: インターフェースはエンティティと同じファイルに定義されていますか？ (例: `internal/domain/trip.go` に `TripRepository` を定義)
-   [ ] **命名規則**: インターフェース名は `<EntityName>Repository` という形式ですか？ (例: `TripRepository`)
-   [ ] **メソッドのシグネチャ**:
    -   [ ] すべてのメソッドの第一引数は `context.Context` ですか？
    -   [ ] メソッドはドメイン固有のエラー（例: `domain.ErrTripNotFound`）または `error` を返しますか？
-   [ ] **引数の型（最重要）**:
    -   [ ] **検索系メソッド** (`FindByID`, `FindByEmail`など) のIDやキー引数は、プリミティブ型 (`string`, `int`) ではなく、**値オブジェクト** (`domain.TripID`, `domain.Email`) になっていますか？
    -   [ ] **操作系メソッド** (`Create`, `Update`, `Delete`) の引数は、IDだけでなく**ドメインエンティティ全体** (`domain.Trip`) を受け取るようになっていますか？
-   [ ] **操作の網羅性**: CRUD (Create, Read, Update, Delete) に対応する基本的なメソッドが定義されていますか？
    -   `Create(ctx context.Context, entity Entity) error`
    -   `FindByID(ctx context.Context, id EntityID) (Entity, error)`
    -   `Update(ctx context.Context, entity Entity) error`
    -   `Delete(ctx context.Context, entity Entity) error`
    -   (必要に応じて) `FindMany`, `FindBy...` など
-   [ ] **モック生成**: `//go:generate mockgen` コメントがインターフェース定義の上に追加されていますか？

---

### 良い実装例

上記のチェックリストに基づいた `TripRepository` の実装例です。

```go
// internal/domain/trip.go の例
//go:generate mockgen -destination mock/trip.go travel-api/internal/domain TripRepository
type TripRepository interface {
	// FindByID は指定されたIDを持つTripエンティティを検索します。
	// 引数にはTripID値オブジェクトを使用し、型安全性を確保します。
	FindByID(ctx context.Context, id TripID) (Trip, error)
	// FindMany はすべてのTripエンティティを検索します。
	FindMany(ctx context.Context) ([]Trip, error)
	// Create は新しいTripエンティティを永続化します。
	// 引数にはTripエンティティ全体を使用します。
	Create(ctx context.Context, trip Trip) error
	// Update は既存のTripエンティティを更新します。
	// 引数には更新対象のTripエンティティ全体を使用します。
	Update(ctx context.Context, trip Trip) error
	// Delete は指定されたTripエンティティを削除します。
	// 引数には削除対象のTripエンティティ全体を使用します。
	Delete(ctx context.Context, trip Trip) error
}
```
`//go:generate mockgen` コメントを追加することで、`go generate` コマンド実行時にテスト用のモック実装が自動生成されます。

### アンチパターン（避けるべき実装）

-   **プリミティブ型を直接使用する**: `string`や`int`などのプリミティブ型をエンティティの識別子として直接引数に取ることは避けてください。例えば、`FindByID(ctx context.Context, id string)` のようにすると、UUID以外の任意の文字列も受け入れてしまい、型安全性が損なわれます。
-   **エンティティの一部のみを引数に取る**: `Delete(ctx context.Context, id string)` のように、ドメインエンティティ全体ではなく、その一部のプリミティブな識別子のみを引数に取ることも避けるべきです。これは、メソッドの意図を曖昧にし、ドメインロジックがリポジトリ層に漏れ出す原因となります。

```go
// Bad Example: internal/domain/trip.go
type TripRepository interface {
    FindByID(ctx context.Context, id string) (Trip, error) // Bad: string型を直接使用
    Delete(ctx context.Context, id string) error           // Bad: string型を直接使用
    // ...
}
```

## 3. エラーの定義

発生する可能性のあるビジネスエラーをドメイン層に定義します。これにより、エラーの種類を明確にし、適切なエラーハンドリングを促します。

-   **エラーコード**: `internal/domain/error_code.go` に定義します。クライアントに返される機械可読なコードです。
-   **エラー構造体**: `internal/domain/error.go` に定義します。エラーコード、開発者向けメッセージ、およびオプションで根本原因を含みます。

**例 (`internal/domain/error_code.go`):**
```go
package domain

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	InternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ValidationError     ErrorCode = "VALIDATION_ERROR"
	TripNotFound        ErrorCode = "TRIP_NOT_FOUND"
)
```

**例 (`internal/domain/error.go`):**
```go
package domain

import "fmt"

// Error はアプリケーション固有のエラーを表すカスタムエラー型です。
type Error struct {
	// Code はクライアントに返される機械可読なエラーコードです。
	Code ErrorCode
	// Message は開発者向けのエラーメッセージです。
	Message string
	// cause はエラーの根本原因です（オプション）。
	cause error
}

// Error はerrorインターフェースを実装します。
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap はエラーチェーンのためにcauseを返します。
func (e *Error) Unwrap() error {
	return e.cause
}

// エラー変数の定義
var (
	// ErrInvalidUUID は、UUIDの形式が無効な場合に返されます。
	ErrInvalidUUID = &Error{Code: ValidationError, Message: "invalid uuid format"}
	// ErrTripNotFound は、Tripが見つからない場合に返されます。
	ErrTripNotFound = &Error{Code: TripNotFound, Message: "trip not found"}
	// ErrInternalServerError は、予期せぬ内部エラーが発生した場合に返されます。
	// このエラーは通常、具体的なエラー情報でラップして使用します。
	ErrInternalServerError = &Error{Code: InternalServerError, Message: "internal server error"}
)

// NewInternalServerError は、具体的なエラー原因を含む内部サーバーエラーを生成します。
func NewInternalServerError(cause error) error {
	return &Error{
		Code:    InternalServerError,
		Message: "internal server error",
		cause:   cause,
	}
}
```

## 4. テスト

ドメイン層のテストは、エンティティの振る舞いやビジネスロジックが期待通りに動作するかを確認するために重要です。

-   **ファイルパス**: `internal/domain/<entity_name>_test.go` (例: `internal/domain/trip_test.go`)
-   **内容**:
    -   エンティティのコンストラクタ関数（例: `NewTrip`）やメソッド（例: `Trip.Update`）の単体テストを記述します。
    -   外部依存（データベースなど）を含まず、純粋なドメインロジックのみをテストします。
    -   `github.com/stretchr/testify/assert` などのアサーションライブラリを使用すると、テストコードを簡潔に記述できます。

```go
// internal/domain/trip_test.go の例
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTripID(t *testing.T) {
	t.Run("正常系: 有効なUUID", func(t *testing.T) {
		validUUID := uuid.New().String()
		tripID, err := NewTripID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, TripID{value: validUUID}, tripID)
	})

	t.Run("異常系: 無効なUUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		tripID, err := NewTripID(invalidUUID)
		assert.ErrorIs(t, err, ErrInvalidUUID)
		assert.Equal(t, TripID{}, tripID)
	})

	t.Run("異常系: 空文字列", func(t *testing.T) {
		emptyUUID := ""
		tripID, err := NewTripID(emptyUUID)
		assert.ErrorIs(t, err, ErrInvalidUUID)
		assert.Equal(t, TripID{}, tripID)
	})
}

func TestNewTrip(t *testing.T) {
	id, err := NewTripID(uuid.New().String())
	assert.NoError(t, err)
	name := "name abc"
	now := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)

	trip := NewTrip(id, name, now, now)

	assert.Equal(t, id, trip.ID)
	assert.Equal(t, name, trip.Name)
	assert.True(t, trip.CreatedAt.Equal(now))
	assert.True(t, trip.UpdatedAt.Equal(now))
}

func TestTrip_Update(t *testing.T) {
	id, err := NewTripID(uuid.New().String())
	assert.NoError(t, err)
	name := "name abc"
	past := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)
	trip := NewTrip(id, name, past, past)

	updatedName := "new name abc"
	now := time.Date(2001, time.January, 1, 1, 1, 1, 1, time.UTC)
	updatedTrip := trip.Update(updatedName, now)

	assert.Equal(t, id, updatedTrip.ID)
	assert.Equal(t, updatedName, updatedTrip.Name)
	assert.True(t, updatedTrip.CreatedAt.Equal(past))
	assert.True(t, updatedTrip.UpdatedAt.Equal(now))
}
```