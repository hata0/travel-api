# インターフェース層の作成

インターフェース層は、外部からのリクエストを受け付け、アプリケーションのユースケースを呼び出し、結果をレスポンスとして返す責務を担います。この層は、Webフレームワーク（Gin）への依存をカプセル化し、以下の主要なコンポーネントで構成されます。

-   **`validator`**: リクエスト（URIパラメータ、JSONボディ）の形式を定義し、バリデーションルールを記述します。
-   **`presenter`**: クライアントに返すJSONレスポンスの構造と、エラーおよび成功レスポンスを生成するためのヘルパーを定義します。
-   **`handler`**: HTTPリクエストを直接処理し、`validator`でリクエストを検証し、`usecase`を呼び出し、`presenter`でレスポンスを構築します。
-   **`middleware`**: 認証など、複数のリクエストにまたがる共通の関心事を処理します。

## 1. バリデーションルールの定義 (`validator`)

リクエストの入力値を検証するための構造体を定義します。

-   **ファイルパス**: `internal/interface/validator/<entity_name>.go`
-   **内容**:
    -   URIパラメータやリクエストボディに対応する構造体を定義します。
    -   `binding` タグを使用して、Ginのバリデーションルールを指定します。

**例 (`trip.go`):**
```go
package validator

// TripURIParameters はURIに含まれるパラメータのバリデーションルールを定義します。
type TripURIParameters struct {
	TripID string `uri:"trip_id" binding:"required"`
}

// CreateTripJSONBody は旅行プラン作成時のリクエストボディのバリデーションルールを定義します。
type CreateTripJSONBody struct {
	Name string `json:"name" binding:"required"`
}

// UpdateTripJSONBody は旅行プラン更新時のリクエストボディのバリデーションルールを定義します。
type UpdateTripJSONBody struct {
	Name string `json:"name" binding:"required"`
}
```

**例 (`auth.go`):**
```go
package validator

type RegisterJSONBody struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginJSONBody struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenJSONBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
```

### バリデーションのテスト

バリデーションルールが期待通りに機能することを確認するために、単体テストを作成します。

-   **ファイルパス**: `internal/interface/validator/<entity_name>_test.go`
-   **内容**:
    -   `github.com/go-playground/validator/v10` を直接使用して、定義した構造体のバリデーションをテストします。
    -   正常系（有効なデータ）と異常系（無効なデータ）の両方のケースをテストします。

**例 (`trip_test.go`):**
```go
package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCreateTripJSONBody_Validation(t *testing.T) {
	validate := validator.New()
	validate.SetTagName("binding")

	t.Run("正常系", func(t *testing.T) {
		params := CreateTripJSONBody{
			Name: "test name",
		}
		err := validate.Struct(params)
		assert.NoError(t, err)
	})

	t.Run("異常系: Nameが空", func(t *testing.T) {
		params := CreateTripJSONBody{
			Name: "",
		}
		err := validate.Struct(params)
		assert.Error(t, err)
	})
}
```

## 2. レスポンス形式の定義 (`presenter`)

クライアントに返すJSONの形式を定義します。成功時とエラー時で共通の構造を提供することで、APIの予測可能性を高めます。

### 2.1. 成功・エラーレスポンス

-   **ファイルパス**: `internal/interface/presenter/success.go`, `internal/interface/presenter/error.go`
-   **内容**:
    -   `success.go`には、成功時のレスポンスを構築するための構造体（`SuccessResponse`）を定義します。これは主に、更新・削除操作が成功し、返すデータがない場合に使用されます。
    -   `error.go`には、エラーレスポンスを統一的に扱うための構造体とファクトリ関数（`ConvertToHTTPError`）を定義します。`ConvertToHTTPError`は、バリデーションエラー、JSONの構文・型エラー、ドメインエラーなど、さまざまな種類のエラーを受け取り、適切なHTTPステータスコードとエラーレスポンスオブジェクトを生成します。内部では `mapErrorCodeToHTTPStatus` を使用して、ドメイン層で定義された `ErrorCode` をHTTPステータスコードにマッピングします。

**例 (`error.go`):**
```go
package presenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"travel-api/internal/domain"

	"github.com/go-playground/validator/v10"
)

// Error はエラーレスポンスの構造を定義します。
type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ConvertToHTTPError はerrorをHTTPステータスコードとレスポンスボディに変換します。
func ConvertToHTTPError(err error) (int, Error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		// バリデーションエラー
		return http.StatusBadRequest, Error{
			Code:    domain.ValidationError.String(),
			Message: "Input validation failed. Please check the details field for more information.",
			Details: formatValidationErrors(validationErrs),
		}
	}

	// ... 他のエラー種別（JSON構文エラー、ドメインエラー等）に対するハンドリング ...

	var appErr *domain.Error
	if errors.As(err, &appErr) {
		// アプリケーションで定義されたドメインエラー
		return mapErrorCodeToHTTPStatus(appErr.Code), Error{
			Code:    appErr.Code.String(),
			Message: appErr.Message,
		}
	}

	// 予期せぬエラー
	slog.Error("An unexpected error occurred", "error", err)
	return http.StatusInternalServerError, Error{
		Code:    domain.InternalServerError.String(),
		Message: "An unexpected internal server error has occurred. Please contact support if the problem persists.",
	}
}
```

### 2.2. データ構造の定義

-   **ファイルパス**: `internal/interface/presenter/<entity_name>.go`
-   **内容**:
    -   APIレスポンスに含めるデータ（例: `Trip`）の構造体と、それを生成するためのコンストラクタ関数を定義します。
    -   必要に応じて`MarshalJSON`をカスタム実装し、`time.Time`型をRFC3339形式の文字列に変換するなど、クライアントが扱いやすい形式にデータを整形します。

**例 (`trip.go`):**
```go
package presenter

import (
	"encoding/json"
	"time"
	"travel-api/internal/usecase/output"
)

type (
	Trip struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	GetTripResponse struct {
		Trip Trip `json:"trip"`
	}

	ListTripResponse struct {
		Trips []Trip `json:"trips"`
	}

	CreateTripResponse struct {
		ID string `json:"id"`
	}
)

func NewGetTripResponse(out output.GetTripOutput) GetTripResponse {
	return GetTripResponse{
		Trip: Trip{
			ID:        out.Trip.ID,
			Name:      out.Trip.Name,
			CreatedAt: out.Trip.CreatedAt,
			UpdatedAt: out.Trip.UpdatedAt,
		},
	}
}

func NewListTripResponse(out output.ListTripOutput) ListTripResponse {
	formattedTrips := make([]Trip, len(out.Trips))
	for i, trip := range out.Trips {
		formattedTrips[i] = Trip{
			ID:        trip.ID,
			Name:      trip.Name,
			CreatedAt: trip.CreatedAt,
			UpdatedAt: trip.UpdatedAt,
		}
	}
	return ListTripResponse{
		Trips: formattedTrips,
	}
}

// MarshalJSON はTrip構造体をJSONにマーシャリングする際のカスタム処理を提供します。
// CreatedAtとUpdatedAtフィールドをRFC3339形式でフォーマットします。
func (t Trip) MarshalJSON() ([]byte, error) {
	type Alias Trip // 無限ループを防ぐためのエイリアス
	return json.Marshal(&struct {
		Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (Alias)(t),
		CreatedAt: t.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339Nano),
	})
}
```

**例 (`auth.go`):**
```go
package presenter

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

type AuthTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
```

## 3. ハンドラの作成 (`handler`)

リクエストを受け取り、ビジネスロジック（ユースケース）を呼び出すコントローラです。

-   **ファイルパス**: `internal/interface/handler/<entity_name>.go`
-   **内容**:
    -   依存するユースケースのインターフェース（例: `TripUsecase`）をフィールドに持ちます。
    -   `RegisterAPI`メソッドで、Ginルーターにエンドポイントを登録します。
    -   Ginのハンドラ関数（例: `get`, `create`）を実装します。
    -   ハンドラ関数内では、以下の処理を順に行います。
        1.  `validator`を使い、リクエストのURIやボディをバインド・検証します。
        2.  ユースケースのメソッドを呼び出します。
        3.  ユースケースの実行結果（出力オブジェクトまたはエラー）を受け取ります。
        4.  `presenter`ヘルパーを使い、成功またはエラーのJSONレスポンスをクライアントに返します。
    -   エラーハンドリングは`presenter.ConvertToHTTPError`に集約され、ハンドラ内のコードをシンプルに保ちます。

**例 (`trip.go`):**
```go
package handler

import (
	"net/http"
	"travel-api/internal/interface/presenter"
	"travel-api/internal/interface/validator"
	"travel-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type TripHandler struct {
	usecase usecase.TripUsecase
}

func (handler *TripHandler) RegisterAPI(router *gin.Engine) {
	router.GET("/trips/:trip_id", handler.get)
	router.GET("/trips", handler.list)
	router.POST("/trips", handler.create)
	router.PUT("/trips/:trip_id", handler.update)
	router.DELETE("/trips/:trip_id", handler.delete)
}

func (handler *TripHandler) get(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.ShouldBindUri(&uriParams); err != nil {
		c.JSON(presenter.ConvertToHTTPError(err)) // 1. バリデーション
		return
	}

	tripOutput, err := handler.usecase.Get(c.Request.Context(), uriParams.TripID) // 2. ユースケース呼び出し
	if err != nil {
		c.JSON(presenter.ConvertToHTTPError(err)) // 3. エラーハンドリング
		return
	}

	// 4. 成功レスポンスの生成
	c.JSON(http.StatusOK, presenter.NewGetTripResponse(tripOutput))
}
```

## 4. 認証ミドルウェア (`middleware`)

認証など、複数のリクエストにまたがる共通の関心事を処理します。

-   **ファイルパス**: `internal/interface/middleware/auth.go`
-   **内容**:
    -   HTTPリクエストの `Authorization` ヘッダーからJWTアクセストークンを抽出します。
    -   トークンの署名と有効性を検証します。
    -   トークンが有効であれば、ユーザーIDをGinのコンテキストに設定し、次のハンドラに処理を渡します。
    -   トークンが無効または欠落している場合は、`presenter.ConvertToHTTPError` を通じて `401 Unauthorized` などのエラーレスポンスを返します。

## 5. テスト

ハンドラのテストは、`net/http/httptest`パッケージを使用してHTTPリクエストをシミュレートし、レスポンスを検証します。

-   **ファイルパス**: `internal/interface/handler/<entity_name>_test.go`
-   **内容**:
    -   `gomock`を使用して、ハンドラが依存するユースケースインターフェースのモックを作成します。
    -   `gin.SetMode(gin.TestMode)` を設定してテストモードでGinを実行します。
    -   `httptest.NewRecorder`でレスポンスを記録し、`http.NewRequest`でリクエストを作成します。
    -   正常系（期待されるレスポンスボディとステータスコードが返るか）と、異常系（バリデーションエラー、ユースケースからのエラーなど）の両方をテストします。

**例 (`trip_test.go`):**
```go
func TestTripHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	NewTripHandler(mockUsecase).RegisterAPI(r)

	t.Run("正常系", func(t *testing.T) {
		// ... モックの期待動作を設定 ...
		mockUsecase.EXPECT().Get(gomock.Any(), "some-id").Return(expectedOutput, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/some-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// ... レスポンスボディの検証 ...
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockUsecase.EXPECT().Get(gomock.Any(), "not-found-id").Return(output.GetTripOutput{}, domain.ErrTripNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/not-found-id", nil)
		r.ServeHTTP(w, req)

		// presenter.ConvertToHTTPError がドメインエラーを適切なHTTPステータスに変換することを検証
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
```
