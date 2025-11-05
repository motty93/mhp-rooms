# 認証アーキテクチャ

このドキュメントでは、mhp-roomsプロジェクトの認証システムのアーキテクチャについて説明します。

## 概要

mhp-roomsではフロントエンド認証アーキテクチャを採用しており、Supabaseの認証機能をフロントエンド（JavaScript）で直接利用し、バックエンド（Go）ではJWTトークンの検証とユーザー情報の管理のみを行います。

## アーキテクチャ図

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   フロントエンド   │    │    Supabase     │    │  バックエンド    │
│   (JavaScript)   │    │    (認証サーバー)  │    │     (Go)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        │
         │ 1. 認証リクエスト         │                        │
         ├────────────────────────→│                        │
         │                        │                        │
         │ 2. JWTトークン           │                        │
         │←────────────────────────┤                        │
         │                        │                        │
         │ 3. API リクエスト (Bearer Token)                  │
         ├─────────────────────────────────────────────────→│
         │                        │                        │
         │                        │ 4. JWT検証              │
         │                        │←───────────────────────┤
         │                        │                        │
         │                        │ 5. 検証結果             │
         │                        ├───────────────────────→│
         │                        │                        │
         │ 6. レスポンス                                     │
         │←─────────────────────────────────────────────────┤
```

## フロントエンド認証処理

### 1. Supabase初期化

**ファイル**: `static/js/supabase.js`

```javascript
// Supabaseクライアントの初期化
window.supabaseClient = window.supabase.createClient(url, anonKey, {
  auth: {
    autoRefreshToken: true,
    persistSession: true,
    detectSessionInUrl: true,
  },
})
```

### 2. 認証フロー

#### ログイン
```javascript
// フロントエンドで直接Supabaseにリクエスト
const { data, error } = await supabase.auth.signInWithPassword({
  email,
  password,
})
```

#### 新規登録
```javascript
const { data, error } = await supabase.auth.signUp({
  email,
  password,
  options: { data: metadata }
})
```

#### パスワードリセット
```javascript
const { data, error } = await supabase.auth.resetPasswordForEmail(email, {
  redirectTo: `${window.location.origin}/auth/reset-password`,
})
```

### 3. セッション管理

**ファイル**: `static/js/auth-store.js` (Alpine.js Store)

```javascript
Alpine.store('auth', {
  user: null,
  session: null,
  loading: true,

  updateSession(session) {
    this.session = session
    this.user = session?.user || null

    // JWTトークンをクッキーに保存（SSR用）
    if (session && session.access_token) {
      document.cookie = `sb-access-token=${session.access_token}; path=/; max-age=3600; SameSite=Lax`
    }
  }
})
```

### 4. API認証ヘッダー設定

**ファイル**: `static/js/htmx-auth.js`

```javascript
// HTMXリクエストに自動でAuthorizationヘッダーを追加
document.body.addEventListener('htmx:beforeRequest', (evt) => {
  if (requestPath.startsWith('/api/') || requestPath.startsWith('/rooms/')) {
    const token = getTokenSync()
    if (token) {
      evt.detail.xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    }
  }
})
```

## バックエンド認証処理

### 1. JWT認証ミドルウェア

**ファイル**: `internal/middleware/auth.go`

```go
func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Authorizationヘッダーまたはクッキーからトークン取得
        token := j.getTokenFromRequest(r)

        // 2. Supabaseでトークン検証
        user, err := j.supabaseClient.Auth.User(ctx, token)

        // 3. アプリケーションDBでユーザー情報取得/作成
        dbUser := j.EnsureUserExists(user)

        // 4. コンテキストにユーザー情報を設定
        ctx = context.WithValue(r.Context(), "user", user)
        ctx = context.WithValue(ctx, "dbUser", dbUser)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. 認証ハンドラーの役割

**ファイル**: `internal/handlers/auth.go`

多くのエンドポイントが「このエンドポイントは使用されません。フロントエンド認証をご利用ください。」というメッセージを返す理由：

#### 🚫 使用されないエンドポイント
- `Login()` - フロントエンドで`supabase.auth.signInWithPassword()`を使用
- `Register()` - フロントエンドで`supabase.auth.signUp()`を使用
- `Logout()` - フロントエンドで`supabase.auth.signOut()`を使用
- `PasswordResetRequest()` - フロントエンドで`supabase.auth.resetPasswordForEmail()`を使用
- `GoogleAuth()` - フロントエンドで`supabase.auth.signInWithOAuth()`を使用

#### ✅ 実際に使用されるエンドポイント
- `SyncUser()` - Supabase認証後にアプリケーションDBにユーザー情報を同期
- `UpdatePSNId()` - PSN IDの更新
- `GetCurrentUser()` - 現在のユーザー情報取得（DBから直接取得）

## ルーティング設定

**ファイル**: `cmd/server/routes.go`

```go
// 認証が不要なページ（オプショナル認証）
r.Get("/", app.withOptionalAuth(ph.Home))
r.Get("/rooms", app.withOptionalAuth(rh.Rooms))

// 認証が必須なページ
r.Get("/profile", app.withAuth(profileHandler.Profile))

// API認証エンドポイント
ar.Post("/auth/sync", app.authHandler.SyncUser)
ar.Put("/auth/psn-id", app.authHandler.UpdatePSNId)
ar.Get("/user/me", app.withAuth(app.authHandler.GetCurrentUser))
```

## 認証フローの詳細

### 1. 新規登録時

```
1. ユーザーが新規登録フォームに入力（メールアドレス、パスワード）
   ※ PSN IDは登録時に入力不要（プロフィール編集で後から設定可能）
2. フロントエンドからSupabaseに新規登録リクエスト
3. Supabaseがアカウントを作成してJWTトークンを発行
4. フロントエンドがJWTトークンをクッキーに保存
5. 認証コールバック画面（/auth/callback）に遷移
6. 自動的に /rooms へリダイレクト
7. /api/auth/sync エンドポイントでアプリケーションDBにユーザー情報を同期
```

### 2. 初回ログイン時

```
1. ユーザーがログインフォームに入力
2. フロントエンドからSupabaseにログインリクエスト
3. Supabaseが認証情報を確認してJWTトークンを発行
4. フロントエンドがJWTトークンをクッキーに保存
5. 自動的に /api/auth/sync エンドポイントを呼び出し
6. バックエンドがSupabaseでトークン検証
7. アプリケーションDBにユーザー情報を作成/更新
```

### 3. 認証が必要なAPI呼び出し時

```
1. フロントエンドがAPIリクエストを送信
2. htmx-auth.jsが自動でAuthorizationヘッダーにJWTトークンを設定
3. バックエンドの認証ミドルウェアがトークンを検証
4. Supabaseでトークンの有効性を確認
5. アプリケーションDBからユーザー情報を取得（キャッシュ利用）
6. リクエストコンテキストにユーザー情報を設定して処理続行
```

### 4. セッション期限切れ時

```
1. 期限切れトークンでAPIリクエスト
2. バックエンドが401 Unauthorizedを返す
3. フロントエンドが401エラーをキャッチ
4. 自動的にログインページにリダイレクト
5. ローカルストレージとクッキーをクリア
```

## 設定とエラーハンドリング

### Supabase設定が未設定の場合

**ファイル**: `static/js/supabase.js`

```javascript
function createDummyAuth() {
  return {
    signIn: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    // ... 他のメソッドも同様
  }
}
```

Supabase設定（`SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_JWT_SECRET`）が未設定の場合、ダミー認証オブジェクトを作成し、すべての認証操作でエラーを返します。

## メリット

1. **セキュリティ**: Supabaseが専門的に認証を処理
2. **開発効率**: 認証ロジックの重複実装が不要
3. **保守性**: 認証機能の更新はSupabase側で自動対応
4. **スケーラビリティ**: 認証処理をSupabaseにオフロード
5. **ユーザビリティ**: クライアントサイドでの高速な認証体験

## 注意点

1. **JavaScript無効環境**: JavaScriptが無効な環境では認証機能が利用できない
2. **Supabase依存**: Supabaseサービスの可用性に依存
3. **設定管理**: Supabase認証設定の適切な管理が必要

## 関連ファイル

### フロントエンド
- `static/js/supabase.js` - Supabase初期化とauth関数
- `static/js/auth-store.js` - Alpine.js認証ストア
- `static/js/htmx-auth.js` - HTMX認証ヘッダー設定

### バックエンド
- `internal/middleware/auth.go` - JWT認証ミドルウェア
- `internal/handlers/auth.go` - 認証関連ハンドラー
- `cmd/server/routes.go` - ルーティング設定

### テンプレート
- `templates/pages/login.tmpl` - ログインページ
- `templates/pages/register.tmpl` - 新規登録ページ
- `templates/pages/auth-callback.tmpl` - OAuth認証コールバック