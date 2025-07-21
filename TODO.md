# TODO

# Auth.goコードレビュー

## 良い点 ✅

### アーキテクチャ設計
- **クリーンアーキテクチャ**: ドメイン駆動設計とレイヤー分離が適切に実装されている
- **依存関係注入**: 外部依存をインターフェースとして抽象化し、テスタビリティが確保されている
- **トランザクション管理**: データベース操作が適切にトランザクションでラップされている

### セキュリティ対策
- **パスワードハッシュ化**: bcryptを使用した適切なパスワード暗号化
- **トークン再利用攻撃対策**: リフレッシュトークンの再利用を検出し、全セッション無効化を実行
- **情報漏洩防止**: ユーザー存在チェックで意図的に曖昧なエラーメッセージを返している
- **JWT適切実装**: アクセストークンに適切な有効期限を設定

### エラーハンドリング
- **ドメインエラー**: 独自エラー型による適切なエラー分類
- **エラーログ**: セキュリティイベントの適切なログ記録

## 改善が必要な点 ⚠️

### 1. トランザクション一貫性の問題
```go
// VerifyRefreshTokenメソッドの問題箇所
_, err := i.revokedTokenRepository.FindByJTI(ctx, refreshToken)
if err == nil {
    // この操作がトランザクション外で実行されている
    foundToken, findErr := i.refreshTokenRepository.FindByToken(ctx, refreshToken)
    if findErr == nil {
        if delErr := i.refreshTokenRepository.DeleteByUserID(ctx, foundToken.UserID); delErr != nil {
            // エラーログのみでロールバックされない
        }
    }
}
```

**修正案**: 全ての操作をトランザクション内で実行する

### 2. リフレッシュトークン設計の脆弱性
現在の実装では、リフレッシュトークンが単純なUUIDです。

**問題点**:
- JWTではないため、サーバーレス環境でのstateless認証ができない
- トークン自体に有効期限情報が含まれていない

**推奨改善**:
```go
// リフレッシュトークンもJWTで実装
refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": userID.String(),
    "type":    "refresh",
    "jti":     i.uuidGenerator.NewUUID(), // JTI for revocation
    "exp":     now.Add(config.RefreshTokenExpiration()).Unix(),
})
```

### 3. セキュリティ強化項目

#### レート制限
```go
// 必要な追加実装
type AuthInteractor struct {
    // 既存フィールド...
    rateLimiter RateLimiter // ログイン試行回数制限
}
```

#### パスワード強度チェック
```go
func (i *AuthInteractor) validatePassword(password string) error {
    if len(password) < 8 {
        return domain.ErrPasswordTooWeak
    }
    // 複雑性チェック、辞書攻撃対策など
    return nil
}
```

### 4. パフォーマンス改善
```go
func (i *AuthInteractor) checkUserExistence(ctx context.Context, username, email string) error {
    // 現在：2回のDB呼び出し
    // 改善案：1回のクエリで両方チェック
    exists, err := i.userRepository.ExistsByUsernameOrEmail(ctx, username, email)
    if err != nil {
        return err
    }
    if exists.UsernameExists {
        return domain.ErrUsernameAlreadyExists
    }
    if exists.EmailExists {
        return domain.ErrEmailAlreadyExists
    }
    return nil
}
```

### 5. 監査ログの追加
```go
func (i *AuthInteractor) Login(ctx context.Context, email, password string) (output.TokenPairOutput, error) {
    start := i.clock.Now()
    defer func() {
        // 成功/失敗に関わらずログインイベントを記録
        i.auditLogger.LogLoginAttempt(ctx, email, start, i.clock.Now())
    }()
    // 既存実装...
}
```

## 実務適用に向けた推奨事項

### 短期改善（即座に対応）
1. **トランザクション範囲の修正**: VerifyRefreshTokenの失効検出処理をトランザクション内に移動
2. **テストカバレッジ**: 特にセキュリティ関連のエッジケースのテスト追加
3. **設定の外部化**: JWT秘密鍵の環境変数化確認

### 中期改善（次回スプリント）
1. **レート制限機能**: Redis/Memcachedを使用したログイン試行制限
2. **パスワード強度チェック**: OWASP推奨レベルの実装
3. **監査ログ**: セキュリティイベントの詳細ログ

### 長期改善（次回バージョン）
1. **JWTリフレッシュトークン**: よりモダンな実装への移行
2. **多要素認証**: TOTP/SMS認証の追加
3. **デバイス管理**: セッション管理の詳細化

## 総合評価

このコードは**実務レベルとして十分に高品質**です。セキュリティの基本要件を満たし、保守しやすい設計になっています。上記の改善点を段階的に適用することで、より堅牢なシステムになるでしょう。

**推奨アクション**: 短期改善項目を優先的に対応し、プロダクション環境での運用を開始して問題ないレベルです。
