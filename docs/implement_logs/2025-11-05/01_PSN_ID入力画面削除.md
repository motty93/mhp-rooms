# PSN ID入力画面削除

## 実装期間
- 開始: 2025-11-05
- 完了: 2025-11-05

## 概要
GitHub issue#67に対応し、新規登録後のPSN ID入力必須画面を削除しました。MHXXやPretendoユーザーなど、PlayStation以外のプラットフォームユーザーも受け入れられるようにするための改善です。

## 実施内容

### 1. 新規登録画面の修正 (`templates/pages/register.tmpl`)
- **削除したもの**:
  - PSN ID入力フィールド（HTML）
  - PSN IDバリデーション処理（JavaScript）
  - フォームデータの`psnId`プロパティ
  - Supabase登録時のメタデータからPSN ID送信

- **結果**:
  - ユーザーはメールアドレスとパスワードのみで登録可能
  - PSN IDは後からプロフィール編集で設定可能

### 2. Complete Profile関連の削除

#### ルーティング (`cmd/server/routes.go`)
- **削除したルート**:
  - `GET /auth/complete-profile` (164行目)
  - `POST /auth/complete-profile` (177行目)

#### ハンドラー (`internal/handlers/auth.go`)
- **削除したメソッド**:
  - `CompleteProfilePage()` (141-146行目)
  - `CompleteProfile()` (148-154行目)

#### テンプレート
- **削除したファイル**:
  - `templates/pages/complete-profile.tmpl`（新しいバージョン）
  - `templates/pages/complete_profile.tmpl`（古いバージョン）

### 3. 認証コールバック処理の修正 (`templates/pages/auth-callback.tmpl`)
- **変更前**: PSN IDの有無をチェックし、未設定の場合は`/auth/complete-profile`へリダイレクト
- **変更後**: PSN IDチェックを削除し、認証成功後は常に`/rooms`へリダイレクト

### 4. 保持した機能
以下の機能は**意図的に保持**しました：
- `static/js/auth-store.js`の`updatePSNId`メソッド
- `internal/handlers/auth.go`の`UpdatePSNId`メソッド
- `PUT /api/auth/psn-id`エンドポイント

**理由**: プロフィール編集画面でPSN IDを後から設定・変更できるようにするため

### 5. ドキュメント更新

#### `docs/routing_design.md`
- Complete Profile関連のルート説明を削除
- 認証コールバックのルート説明を追加

#### `docs/api-design.md`
- Complete Profile関連のAPIエンドポイント説明を削除
- 認証ページ表示エンドポイントを更新

#### `docs/authentication-architecture.md`
- 新規登録フローの説明を追加
- PSN IDが登録時に不要であることを明記
- 認証フローの番号を修正（重複していた「2.」を修正）

## 新しい認証フロー

### 変更前
```
1. 新規登録画面でメール、パスワード、PSN IDを入力
2. Supabase登録（メタデータにPSN ID含む）
3. 認証コールバック
4. PSN IDチェック
   - 未設定 → /auth/complete-profile へリダイレクト（PSN ID入力を強制）
   - 設定済み → /rooms へリダイレクト
```

### 変更後
```
1. 新規登録画面でメール、パスワードのみを入力
2. Supabase登録
3. 認証コールバック
4. /rooms へリダイレクト
5. PSN IDは後からプロフィール編集で任意に設定可能
```

## 影響範囲

### 変更されたファイル
- `templates/pages/register.tmpl`
- `templates/pages/auth-callback.tmpl`
- `cmd/server/routes.go`
- `internal/handlers/auth.go`
- `docs/routing_design.md`
- `docs/api-design.md`
- `docs/authentication-architecture.md`

### 削除されたファイル
- `templates/pages/complete-profile.tmpl`
- `templates/pages/complete_profile.tmpl`

## テスト確認項目

### 新規登録フロー
- [ ] メールアドレスとパスワードのみで登録できること
- [ ] 登録後、/roomsにリダイレクトされること
- [ ] Complete Profile画面にリダイレクトされないこと

### Google OAuth認証フロー
- [ ] Google認証後、/roomsにリダイレクトされること
- [ ] PSN IDチェックが行われないこと

### プロフィール編集
- [ ] プロフィール編集画面でPSN IDを設定できること
- [ ] PSN IDを変更できること

### エンドポイント削除確認
- [ ] GET /auth/complete-profile が404になること
- [ ] POST /auth/complete-profile が404になること

## 今後の改善点

### プロフィール編集画面の改善
現在、PSN IDはプロフィール編集で設定可能ですが、以下の改善が考えられます：
- プラットフォームIDの種類を増やす（Nintendo ID、Xbox Gamertag等）
- 複数のプラットフォームIDを登録できるようにする
- 各プラットフォームの表示アイコンを追加

### ユーザーオンボーディング
新規登録後、初回ログイン時に以下のガイドを表示することを検討：
- プロフィール設定の促進（任意）
- 使い方ガイドへのリンク
- プラットフォームID設定の案内

## 関連issue
- GitHub issue #67: 新規登録後のPSN ID入力画面を無くす

## 特記事項
- 既存ユーザーのPSN IDデータは保持されます
- PSN ID未設定ユーザーも問題なくサービスを利用できます
- プロフィール編集機能でPSN IDを後から設定可能です
