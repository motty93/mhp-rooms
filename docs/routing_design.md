# URL構造とルーティング設計

## 概要

HuntersHubのURL構造とルーティングの設計ドキュメントです。SEO対策と使いやすさを両立させた設計を目指します。

## 現在の実装状況

### 実装済みのルート

#### 1. 基本ページ
```
/                           # ホームページ（ランディングページ）
/terms                      # 利用規約
/privacy                    # プライバシーポリシー
/contact                    # お問い合わせ
/faq                        # よくある質問
/guide                      # 使い方ガイド
/sitemap.xml                # サイトマップ
```

#### 2. 認証関連
```
/auth/login                 # ログインページ
/auth/register              # 新規登録ページ
/auth/logout                # ログアウト（POST）
/auth/password-reset        # パスワードリセット
/auth/password-reset/confirm # パスワードリセット確認
/auth/complete-profile      # プロフィール補完
/auth/google                # Google OAuth認証（準備中）
/auth/google/callback       # Google OAuth コールバック
```

#### 3. ルーム関連
```
/rooms                      # ルーム一覧
/rooms                      # ルーム作成（POST）
/rooms/{id}/join            # ルーム参加（POST）
/rooms/{id}/leave           # ルーム退出（POST）
/rooms/{id}/toggle-closed   # ルーム開閉切替（PUT）
```

#### 4. API エンドポイント
```
/api/user/current           # 現在のユーザー情報
/hello                      # テスト用エンドポイント
/health                     # ヘルスチェック
```

### ルーティング実装（Chi）

現在の`cmd/server/main.go`での実装：

```go
r := chi.NewRouter()

// 基本ページ
r.HandleFunc("/", h.HomeHandler).Methods("GET")
r.HandleFunc("/rooms", h.RoomsHandler).Methods("GET")
r.HandleFunc("/rooms", h.CreateRoomHandler).Methods("POST")
r.HandleFunc("/rooms/{id}/join", h.JoinRoomHandler).Methods("POST")
r.HandleFunc("/rooms/{id}/leave", h.LeaveRoomHandler).Methods("POST")
r.HandleFunc("/rooms/{id}/toggle-closed", h.ToggleRoomClosedHandler).Methods("PUT")

// 法的ページ
r.HandleFunc("/terms", h.TermsHandler).Methods("GET")
r.HandleFunc("/privacy", h.PrivacyHandler).Methods("GET")
r.HandleFunc("/contact", h.ContactHandler).Methods("GET", "POST")
r.HandleFunc("/faq", h.FAQHandler).Methods("GET")
r.HandleFunc("/guide", h.GuideHandler).Methods("GET")

// 認証関連
r.HandleFunc("/auth/login", h.LoginPageHandler).Methods("GET")
r.HandleFunc("/auth/login", h.LoginHandler).Methods("POST")
r.HandleFunc("/auth/register", h.RegisterPageHandler).Methods("GET")
r.HandleFunc("/auth/register", h.RegisterHandler).Methods("POST")
r.HandleFunc("/auth/logout", h.LogoutHandler).Methods("POST")

// パスワードリセット
r.HandleFunc("/auth/password-reset", h.PasswordResetPageHandler).Methods("GET")
r.HandleFunc("/auth/password-reset", h.PasswordResetRequestHandler).Methods("POST")
r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirmPageHandler).Methods("GET")
r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirmHandler).Methods("POST")

// Google OAuth
r.HandleFunc("/auth/google", h.GoogleAuthHandler).Methods("GET")
r.HandleFunc("/auth/google/callback", h.GoogleCallbackHandler).Methods("GET")

// プロフィール補完
r.HandleFunc("/auth/complete-profile", h.CompleteProfilePageHandler).Methods("GET")
r.HandleFunc("/auth/complete-profile", h.CompleteProfileHandler).Methods("POST")

// API
r.HandleFunc("/api/user/current", h.CurrentUserHandler).Methods("GET")

// その他
r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")
r.HandleFunc("/sitemap.xml", h.SitemapHandler).Methods("GET")
r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}).Methods("GET")

// 静的ファイル
r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
```

## 問題点と改善提案

### 1. ミドルウェアが未適用
**問題**: 定義されているミドルウェアが実際には使用されていない
- `AuthMiddleware`: 認証状態の確認
- `RequireAuthMiddleware`: 認証必須エンドポイントの保護
- `ProfileCompleteMiddleware`: プロフィール完成チェック

**改善案**:
```go
// 認証が必要なルート
authRequired := r.PathPrefix("/").Subrouter()
authRequired.Use(h.AuthMiddleware)
authRequired.Use(h.RequireAuthMiddleware)

authRequired.HandleFunc("/rooms", h.CreateRoomHandler).Methods("POST")
authRequired.HandleFunc("/rooms/{id}/join", h.JoinRoomHandler).Methods("POST")
authRequired.HandleFunc("/rooms/{id}/leave", h.LeaveRoomHandler).Methods("POST")
authRequired.HandleFunc("/rooms/{id}/toggle-closed", h.ToggleRoomClosedHandler).Methods("PUT")

// プロフィール完成が必要なルート
profileRequired := authRequired.PathPrefix("/").Subrouter()
profileRequired.Use(h.ProfileCompleteMiddleware)
profileRequired.HandleFunc("/rooms", h.CreateRoomHandler).Methods("POST")
```

### 2. マルチプラットフォーム対応の準備

現在はPSP専用の実装ですが、将来的なマルチプラットフォーム対応に向けた設計：

```
# 将来の拡張案
/psp/rooms                  # PSP専用ルーム
/3ds/rooms                  # 3DS専用ルーム
/wiiu/rooms                 # Wii U専用ルーム
/rooms?platform=psp         # フィルタリング対応
```

### 3. RESTful API の整備

現在のAPIは限定的なので、より体系的なAPIエンドポイントの実装：

```
/api/v1/rooms               # ルーム一覧（GET）
/api/v1/rooms               # ルーム作成（POST）
/api/v1/rooms/{id}          # ルーム詳細（GET）
/api/v1/rooms/{id}          # ルーム更新（PUT）
/api/v1/rooms/{id}          # ルーム削除（DELETE）
/api/v1/rooms/{id}/members  # メンバー一覧（GET）
/api/v1/user/profile        # プロフィール取得（GET）
/api/v1/user/profile        # プロフィール更新（PUT）
```

## セキュリティ考慮事項

### 1. CSRF対策
- Supabaseの認証トークンを使用
- SameSiteクッキー属性の活用

### 2. 認証保護
- ルーム作成、参加、退出は認証必須に
- プロフィール未完成ユーザーの機能制限

### 3. レート制限（未実装）
- ルーム作成の頻度制限
- API呼び出しの制限

## パフォーマンス最適化

### 1. 静的ファイル
- `/static/`パスで効率的に配信
- 将来的にはCDN対応を検討

### 2. キャッシュ戦略
- 静的ページのキャッシュヘッダー設定
- 動的コンテンツの適切なキャッシュ制御

## 実装優先度

1. **高優先度**（セキュリティ関連）
   - [ ] ミドルウェアの適用
   - [ ] 認証必須エンドポイントの保護

2. **中優先度**（機能改善）
   - [ ] RESTful APIの整備
   - [ ] エラーページ（404, 500）の実装

3. **低優先度**（将来の拡張）
   - [ ] マルチプラットフォーム対応
   - [ ] レート制限の実装
   - [ ] 国際化対応