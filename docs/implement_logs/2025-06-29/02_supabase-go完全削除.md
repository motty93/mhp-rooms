# Supabase-go完全削除

**実装開始時刻**: 2025-06-29 (開始時刻不明)  
**実装完了時刻**: 2025-06-29 (完了時刻不明)  
**実装時間**: 約30分

## 実装概要

Supabase-jsによるフロントエンド認証への完全移行に伴い、サーバーサイドで使用していたsupabase-goの依存関係を完全に削除しました。これにより、認証処理が完全にフロントエンド主導となり、サーバーはJWT検証のみを行うシンプルな構成になりました。

## 削除した内容

### 1. Supabaseクライアント関連
- **ディレクトリ**: `/internal/infrastructure/auth/supabase/`
  - `client.go` - Supabaseクライアントの初期化と管理
- **アプリケーション初期化**: `cmd/server/application.go`
  - Supabase初期化処理を削除

### 2. ハンドラーからの依存削除
- **BaseHandler構造体** (`internal/handlers/handlers.go`)
  - `supabase *supa.Client` フィールドを削除
  - supabase-goのインポートを削除

- **各ハンドラーのコンストラクタ**
  - `NewAuthHandler()`, `NewRoomHandler()`, `NewPageHandler()`
  - Supabaseクライアント引数を削除

### 3. 認証関連ファイル
- **削除したファイル**:
  - `internal/handlers/google_auth.go`
  - `internal/handlers/google_auth_supabase.go`
  - `internal/handlers/middleware.go` (旧認証ミドルウェア)
  - `internal/handlers/password_reset.go`
  - `internal/handlers/auth_test.go`

### 4. 認証メソッドの無効化
- **AuthHandler** (`internal/handlers/auth.go`)
  - `Login()`, `Register()`, `Logout()` メソッドを無効化
  - `PasswordResetRequest()`, `PasswordResetConfirm()` メソッドを無効化
  - `GoogleAuth()`, `GoogleCallback()` メソッドを無効化
  - 各メソッドは501 Not Implementedを返すように変更

### 5. 依存関係の削除
- **go.mod**から以下のパッケージが自動削除:
  - `github.com/supabase-community/supabase-go`
  - `github.com/supabase-community/gotrue-go`
  - `github.com/supabase-community/functions-go`
  - `github.com/supabase-community/postgrest-go`
  - `github.com/supabase-community/storage-go`

## 保持した機能

### 1. ページレンダリング
- ログイン・新規登録ページの表示機能は維持
- パスワードリセットページの表示機能は維持

### 2. JWT認証システム
- **JWT検証ミドルウェア** (`internal/middleware/auth.go`)
  - 独自実装のJWT検証システムは保持
  - Supabase JWTトークンの検証は継続

### 3. 設定エンドポイント
- **ConfigHandler** (`internal/handlers/config.go`)
  - フロントエンド用Supabase設定の提供

## 変更の影響

### 1. **認証フロー**
- **Before**: サーバーサイドでSupabase APIを呼び出し
- **After**: フロントエンドでSupabase-jsを使用、サーバーはJWT検証のみ

### 2. **API エンドポイント**
- 既存の認証エンドポイント（`/auth/login`, `/auth/register`等）は501エラーを返す
- `/api/config/supabase`エンドポイントでフロントエンド設定を提供

### 3. **コードベースの簡素化**
- 依存関係が大幅に減少
- 認証に関するサーバーサイドロジックが大幅に簡素化

## テスト結果

### ✅ 成功項目
- **ビルドテスト**: `go build ./cmd/server` 成功
- **依存関係**: `go mod tidy` でsupabase関連パッケージが正常に削除
- **テスト実行**: 全テストパッケージが正常に実行（テストファイルなし状態）

### 🔧 調整が必要な項目
- 削除した認証エンドポイントを参照している既存のテンプレートやJavaScript
- 新しいフロントエンド認証フローでの実際の動作テスト

## 今後の作業

1. **動作確認**
   - 実際のSupabaseプロジェクトでの動作テスト
   - フロントエンド認証フローの確認

2. **ドキュメント更新**
   - 認証方式変更に関するドキュメント更新
   - 開発者向けガイドの更新

3. **エラーハンドリング改善**
   - 無効化されたエンドポイントのより適切なエラーレスポンス

## まとめ

Supabase-goの完全削除により、認証システムが完全にフロントエンド主導となりました。これにより：

- **シンプルな構成**: サーバーはJWT検証のみに集中
- **軽量化**: 依存関係の大幅削減
- **現代的**: フロントエンド認証によるUX向上

この変更により、保守性とパフォーマンスの向上が期待されます。