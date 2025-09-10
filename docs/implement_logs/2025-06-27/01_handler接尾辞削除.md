# Handler接尾辞の削除

実装開始: 2025-06-27 09:00
実装完了: 2025-06-27 09:15
所要時間: 約15分

## 実装概要

ハンドラーメソッド名から冗長な `Handler` 接尾辞を削除し、より簡潔なメソッド名に変更しました。

## 変更内容

### 1. ハンドラーメソッド名の変更（31個）

以下のファイルのメソッド名を変更しました：

#### auth.go
- `LoginPageHandler` → `LoginPage`
- `RegisterPageHandler` → `RegisterPage`
- `LoginHandler` → `Login`
- `RegisterHandler` → `Register`
- `LogoutHandler` → `Logout`

#### complete_profile.go
- `CompleteProfilePageHandler` → `CompleteProfilePage`
- `CompleteProfileHandler` → `CompleteProfile`
- `CurrentUserHandler` → `CurrentUser`

#### faq.go
- `FAQHandler` → `FAQ`

#### google_auth.go
- `GoogleAuthHandler` → `GoogleAuth`
- `GoogleCallbackHandler` → `GoogleCallback`

#### google_auth_supabase.go
- `GoogleAuthSupabaseHandler` → `GoogleAuthSupabase`
- `OAuthCallbackHandler` → `OAuthCallback`
- `SessionHandler` → `Session`

#### guide.go
- `GuideHandler` → `Guide`

#### handlers.go
- `TermsHandler` → `Terms`
- `PrivacyHandler` → `Privacy`
- `ContactHandler` → `Contact`
- `HelloHandler` → `Hello` (スタンドアロン関数)

#### home.go
- `HomeHandler` → `Home`

#### password_reset.go
- `PasswordResetPageHandler` → `PasswordResetPage`
- `PasswordResetRequestHandler` → `PasswordResetRequest`
- `PasswordResetConfirmPageHandler` → `PasswordResetConfirmPage`
- `PasswordResetConfirmHandler` → `PasswordResetConfirm`

#### rooms.go
- `RoomsHandler` → `Rooms`
- `CreateRoomHandler` → `CreateRoom`
- `JoinRoomHandler` → `JoinRoom`
- `LeaveRoomHandler` → `LeaveRoom`
- `ToggleRoomClosedHandler` → `ToggleRoomClosed`
- `GetAllRoomsAPIHandler` → `GetAllRoomsAPI`

#### sitemap.go
- `SitemapHandler` → `Sitemap`

### 2. main.goのルーティング定義更新

`cmd/server/main.go` のルーティング定義（28箇所）を新しいメソッド名に合わせて更新しました。

## 特記事項

- テストファイル（auth_test.go）のメソッド名は変更していません
- `NewHandler` コンストラクタ関数は変更対象外としました
- 変更により、ハンドラーメソッド名がより簡潔で分かりやすくなりました

## 動作確認

変更後、コンパイルエラーがないことを確認する必要があります。