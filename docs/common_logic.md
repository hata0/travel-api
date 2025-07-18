# 共通ロジック

このドキュメントでは、アプリケーション全体で利用される共通のロジックやユーティリティについて説明します。
これらは特定のドメインに依存せず、再利用可能な形で提供されます。

## 時刻管理 (`clock.go`)

`internal/domain/clock.go` では、現在時刻を取得するための抽象化を提供します。
これにより、テスト時に時刻をモックしたり、異なる時刻ソースを容易に切り替えたりすることが可能になります。

### `Clock` インターフェース

現在時刻を返す `Now()` メソッドを定義します。

```go
// Clock は現在時刻を取得するためのインターフェースです。
type Clock interface {
	Now() time.Time
}
```

### `SystemClock` 実装

`Clock` インターフェースのデフォルト実装で、システムの現在時刻を返します。

```go
// SystemClock はClockのデフォルト実装です。
type SystemClock struct{}

// Now は現在時刻を返します。
func (c *SystemClock) Now() time.Time {
	return time.Now()
}
```

## UUID生成とバリデーション (`uuid.go`)

`internal/domain/uuid.go` では、UUIDの生成とバリデーションに関する機能を提供します。

### `UUIDGenerator` インターフェース

新しいUUID文字列を生成する `NewUUID()` メソッドを定義します。

```go
// UUIDGenerator はUUIDを生成するためのインターフェースです。
type UUIDGenerator interface {
	NewUUID() string
}
```

### `DefaultUUIDGenerator` 実装

`UUIDGenerator` インターフェースのデフォルト実装で、`github.com/google/uuid` ライブラリを使用して新しいUUIDを生成します。

```go
// DefaultUUIDGenerator はUUIDGeneratorのデフォルト実装です。
type DefaultUUIDGenerator struct{}

// NewUUID は新しいUUIDを生成します。
func (g *DefaultUUIDGenerator) NewUUID() string {
	return uuid.New().String()
}
```

### `IsValidUUID` 関数

与えられた文字列が有効なUUIDv4形式であるかを検証するヘルパー関数です。

```go
func IsValidUUID(id string) bool {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	return parsedUUID.Version() == 4
}
```
