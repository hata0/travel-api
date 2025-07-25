# ユースケース層の作成

ユースケース層は、アプリケーション固有のビジネスルールを定義し、ドメイン層とインターフェース層の間の調整役を担います。
ここでは、特定のユーザー操作やシステムイベントに対応するアプリケーションの振る舞いを記述します。

## 1. ユースケースインターフェースの定義

まず、ユースケースが提供する機能をインターフェースとして定義します。これにより、インターフェース層（ハンドラ）は具体的な実装に依存せず、このインターフェースにのみ依存することになります。

-   **ファイルパス**: `internal/usecase/<entity_name>.go` (例: `internal/usecase/trip.go`)
-   **内容**:
    -   CRUD操作など、ハンドラから呼び出される一連のメソッドを定義します。
    -   `//go:generate mockgen`コメントを追加することで、`go generate`コマンド実行時にテスト用のモック実装が自動生成されます。

```go
// internal/usecase/trip.go
//go:generate mockgen -destination mock/trip.go travel-api/internal/usecase TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (output.GetTripOutput, error)
	List(ctx context.Context) (output.ListTripOutput, error)
	Create(ctx context.Context, name string) (string, error)
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}
```

## 2. インタラクターの定義

次に、ユースケースインターフェースを実装するインタラクターを定義します。

-   **ファイルパス**: `internal/usecase/<entity_name>.go` (例: `internal/usecase/trip.go`)
-   **内容**:
    -   ユースケースインターフェースを実装する構造体（例: `TripInteractor`）を定義します。
    -   ドメイン層のリポジトリインターフェースや、時刻、UUID生成などの外部依存をコンストラクタを通じて注入します。
    -   **トランザクションを必要とするユースケース（例: 認証関連）では、`domain.TransactionManager`も注入されます。** これにより、複数のデータベース操作をアトミックに実行できます。
    -   各メソッドは、入力値の検証（必要であれば）、ドメインオブジェクトの操作、リポジトリを通じた永続化、そして出力値の生成を行います。
    -   ビジネスロジックの調整役として機能し、ドメイン層のエンティティやリポジトリを直接操作します。

```go
// internal/usecase/trip.go の例
type TripInteractor struct {
	repository    domain.TripRepository
	clock         domain.Clock
	uuidGenerator domain.UUIDGenerator
}

func NewTripInteractor(repository domain.TripRepository, clock domain.Clock, uuidGenerator domain.UUIDGenerator) *TripInteractor {
	return &TripInteractor{
		repository:    repository,
		clock:         clock,
		uuidGenerator: uuidGenerator,
	}
}

// internal/usecase/auth.go の例 (TransactionManager の注入)
type AuthInteractor struct {
	userRepository         domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	clock                  domain.Clock
	uuidGenerator          domain.UUIDGenerator
	transactionManager     domain.TransactionManager // TransactionManager を注入
}

func NewAuthInteractor(userRepository domain.UserRepository, refreshTokenRepository domain.RefreshTokenRepository, clock domain.Clock, uuidGenerator domain.UUIDGenerator, transactionManager domain.TransactionManager) *AuthInteractor {
	return &AuthInteractor{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		clock:                  clock,
		uuidGenerator:          uuidGenerator,
		transactionManager:     transactionManager,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (output.GetTripOutput, error) {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return output.GetTripOutput{}, err
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		return output.GetTripOutput{}, err
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *AuthInteractor) Login(ctx context.Context, email, password string) (output.TokenPairOutput, error) {
	var tokenPair output.TokenPairOutput
	// transactionManager.RunInTx を使用して、複数のDB操作をアトミックに実行
	err := i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error {
		// ... ログインロジック ...
		// リポジトリメソッドには txCtx を渡すことで、トランザクション内で動作する
		user, err := i.userRepository.FindByEmail(txCtx, email)
		// ...
		err = i.refreshTokenRepository.Create(txCtx, newRefreshToken)
		// ...
		return nil
	})

	return tokenPair, err
}

// ... 他のメソッド (Create, Update, Delete, VerifyRefreshToken, RevokeRefreshToken) の実装 ...
```

## 3. 入出力の定義

ユースケースの入出力は、シンプルで具体的なデータ構造として定義します。これにより、ユースケースのインターフェースが明確になり、依存関係が整理されます。

-   **ファイルパス**: `internal/usecase/output/<entity_name>.go` (例: `internal/usecase/output/trip.go`)
-   **内容**:
    -   ユースケースの出力データを表す構造体（例: `GetTripOutput`, `ListTripOutput`）を定義します。
    -   ドメイン層のエンティティを直接公開せず、ユースケース層の関心事に合わせた形式に変換します。

```go
// internal/usecase/output/trip.go の例
package output

import (
	"time"
	"travel-api/internal/domain"
)

type Trip struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetTripOutput struct {
	Trip Trip
}

func NewGetTripOutput(trip domain.Trip) GetTripOutput {
	return GetTripOutput{
		Trip: mapToTrip(trip),
	}
}

type ListTripOutput struct {
	Trips []Trip
}

func NewListTripOutput(trips []domain.Trip) ListTripOutput {
	formattedTrips := make([]Trip, len(trips))
	for i, trip := range trips {
		formattedTrips[i] = mapToTrip(trip)
	}

	return ListTripOutput{
		Trips: formattedTrips,
	}
}

func mapToTrip(trip domain.Trip) Trip {
	return Trip{
		ID:        trip.ID.String(),
		Name:      trip.Name,
		CreatedAt: trip.CreatedAt,
		UpdatedAt: trip.UpdatedAt,
	}
}
```

## 4. テスト

ユースケース層のテストは、ビジネスロジックが期待通りに動作するかを確認するために重要です。外部依存はモック化し、純粋なユースケースの振る舞いをテストします。

-   **ファイルパス**: `internal/usecase/<entity_name>_test.go` (例: `internal/usecase/trip_test.go`)
-   **内容**:
    -   `go.uber.org/mock/gomock` を使用して、`domain.TripRepository`、`domain.Clock`、`domain.UUIDGenerator` などの依存をモック化します。
    -   **トランザクションを伴うユースケースのテストでは、`domain.TransactionManager`もモック化し、`RunInTx`メソッドが渡された関数を実行するように設定します。**
    -   各ユースケースメソッドの正常系と異常系（例: 無効な入力、リポジトリからのエラー）を網羅的にテストします。
    -   `github.com/stretchr/testify/assert` などのアサーションライブラリを使用すると、テストコードを簡潔に記述できます。

```go
// internal/usecase/trip_test.go の例
func TestTripInteractor_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	tripName := "New Trip"
	generatedUUID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	mockUUIDGenerator.EXPECT().NewUUID().Return(generatedUUID).Times(1)
	mockClock.EXPECT().Now().Return(now).Times(2)

	tripID, err := domain.NewTripID(generatedUUID)
	assert.NoError(t, err)
	expectedTrip := domain.NewTrip(tripID, tripName, now, now)

	mockRepo.EXPECT().Create(gomock.Any(), expectedTrip).Return(nil).Times(1)

	createdID, err := interactor.Create(context.Background(), tripName)

	assert.NoError(t, err)
	assert.Equal(t, generatedUUID, createdID)
}

// internal/usecase/auth_test.go の例 (TransactionManager のモック化)
func TestAuthInteractor_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_domain.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mock_domain.NewMockRefreshTokenRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	mockTransactionManager := mock_domain.NewMockTransactionManager(ctrl)
	interactor := NewAuthInteractor(mockUserRepo, mockRefreshTokenRepo, mockClock, mockUUIDGenerator, mockTransactionManager)

	// RunInTx メソッドが渡された関数をそのまま実行するように設定
	mockTransactionManager.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	).AnyTimes() // 複数回呼び出される可能性があるため AnyTimes() を使用

	// ... その他のモック設定とテストロジック ...

	// 正常系: ユーザーが正常にログインできる
	t.Run("正常系: ユーザーが正常にログインできる", func(t *testing.T) {
		// ... FindByEmail, NewUUID, Now, Create などのモック設定 ...
		// リポジトリメソッドの EXPECT には gomock.Any() を使用し、txCtx を考慮しない
		mockUserRepo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, nil).AnyTimes()
		mockRefreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		output, err := interactor.Login(context.Background(), "test@example.com", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, output.Token)
	})

	// ... その他のテストケース ...
}

func TestTripInteractor_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_domain.NewMockClock(ctrl)
	mockUUIDGenerator := mock_domain.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	t.Run("正常系: Tripが取得できる", func(t *testing.T) {
		tripID, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		now := time.Now()
		expectedTrip := domain.NewTrip(tripID, "Test Trip", now, now)

		mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(expectedTrip, nil).Times(1)

		tripOutput, err := interactor.Get(context.Background(), tripID.String())

		assert.NoError(t, err)
		assert.Equal(t, output.NewGetTripOutput(expectedTrip), tripOutput)
	})

	// ... その他のテストケース ...
}
```