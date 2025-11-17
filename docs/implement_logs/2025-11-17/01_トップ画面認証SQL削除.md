# トップ画面 認証とSQLの発行を削除

**実装開始**: 2025-11-17
**実装完了**: 2025-11-17
**所要時間**: 約30分
**関連issue**: #151

## 概要

トップページ（/）から認証制御と認証関連のJavaScriptを削除し、SQLが一切発行されないように修正しました。これにより、ページ読み込みの高速化とサーバー負荷の軽減を実現しました。

## 背景

トップ画面はサービスのファーストビューであり、わざわざ認証によって制御をすべきでないと判断しました。

### 以前の問題点

1. **不要な認証チェック**: トップページでオプショナル認証ミドルウェアを使用
2. **SQLの発行**: 認証済みユーザーがアクセスすると以下のAPIが自動的に呼ばれる
   - `/api/auth/sync` - ユーザー同期
   - `/api/user/me` - DBユーザー情報取得
   - `/api/user/current-room` - 現在の部屋情報取得
3. **不要なリダイレクト**: 認証済みユーザーを自動的に `/rooms` にリダイレクト
4. **ページフラッシュ**: JavaScript実行によるリダイレクトで一瞬トップページが表示される

## 修正内容

### 1. ルーティングの変更

**ファイル**: `cmd/server/routes.go`

**変更箇所**: 85行目

```go
// 変更前
r.Get("/", app.withOptionalAuth(ph.Home))

// 変更後
r.Get("/", ph.Home)
```

**変更点**: トップページから `withOptionalAuth` ミドルウェアを削除

---

### 2. ハンドラーの変更

**ファイル**: `internal/handlers/home.go`

**変更箇所**: 5-11行目

```go
// 変更前
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
    data := TemplateData{
        Title:   "ホーム",
        HasHero: true,
    }
    renderTemplate(w, "home.tmpl", data)
}

// 変更後
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
    data := TemplateData{
        Title:      "ホーム",
        HasHero:    true,
        StaticPage: true,  // 認証関連JSの読み込みを無効化
    }
    renderTemplate(w, "home.tmpl", data)
}
```

**変更点**: `StaticPage: true` を設定することで、`base.tmpl` で認証関連のJavaScriptが読み込まれなくなる

---

### 3. テンプレートの変更

**ファイル**: `templates/pages/home.tmpl`

#### 3-1. 認証チェック＆リダイレクトJavaScriptの削除

**削除箇所**: 34-47行目（変更前の行番号）

```javascript
// 削除したコード
<script>
  document.addEventListener('DOMContentLoaded', function () {
    setTimeout(() => {
      if (window.Alpine && window.Alpine.store('auth')) {
        const auth = window.Alpine.store('auth')
        if (auth.isAuthenticated) {
          window.location.href = '/rooms'
        }
      }
    }, 100)
  })
</script>
```

**理由**: 認証済みユーザーもトップページをそのまま表示する仕様に変更

#### 3-2. 部屋作成ボタンの削除

**削除箇所**: 76-94行目（変更前の行番号）

```html
<!-- 削除したコード -->
<!-- 認証済みユーザー用の部屋作成ボタン -->
<button
  @click="$store.roomCreate.open()"
  x-show="$store.auth.isAuthenticated"
  style="display: none;"
  class="bg-white bg-opacity-20 hover:bg-opacity-30 text-white font-medium py-3 px-8 rounded-lg transition-colors backdrop-blur-sm"
>
  部屋を作る
</button>

<!-- 未認証ユーザー用の無効化ボタン -->
<button
  x-show="!$store.auth.isAuthenticated"
  @click="$store.auth.handleUnauthenticatedAction()"
  class="bg-gray-400 bg-opacity-50 text-gray-300 font-medium py-3 px-8 rounded-lg cursor-not-allowed backdrop-blur-sm"
  title="ログインが必要です"
>
  部屋を作る
</button>
```

**残したコード**: 「部屋を見る」リンクのみ

```html
<div class="flex flex-col sm:flex-row gap-4 justify-center">
  <a
    href="/rooms"
    class="text-white hover:bg-white hover:bg-opacity-20 font-medium py-3 px-8 rounded-lg transition-colors inline-block text-center"
  >
    部屋を見る
  </a>
</div>
```

**理由**: 部屋作成ボタンは完全に削除する仕様

### 4. ヘッダーの変更

**ファイル**: `templates/components/header.tmpl`

**変更点**: `StaticPage` フラグが有効な場合は、Alpine.js のストアに依存しない静的なヘッダーを描画します。ロゴリンクは `/` 固定、右側は未認証時と同じ「ログイン」「新規登録」ボタンのみを常時表示し、プロフィールアイコンやドロップダウン、部屋作成ボタンなどの認証UIは描画しません。

**理由**: トップページで認証用JavaScriptを読み込まなくても期待通りのヘッダーが表示されるようにし、issueで求められていた「ヘッダーのプロフィールアイコンの認証制御を無くす」仕様を満たすため。

---

### 4. 確認のみ（変更不要）

**ファイル**: `templates/layouts/base.tmpl`

**確認箇所**: 99-112行目

```html
{{ if not .StaticPage }}
  <!-- Supabase設定 -->
  <script>
    window.SUPABASE_CONFIG = {
      url: '{{ getEnv "SUPABASE_URL" "" }}',
      anonKey: '{{ getEnv "SUPABASE_ANON_KEY" "" }}'
    };
  </script>
  <script src="https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2"></script>
  <script src="/static/js/supabase.js"></script>
  <script src="/static/js/auth-store.js"></script>
  <script src="/static/js/room-create-store.js"></script>
  <script src="/static/js/htmx-auth.js"></script>
{{ end }}
```

**理由**: 既に `StaticPage` フラグで条件分岐されているため、変更不要

---

## 期待される効果

### ✅ SQL発行の完全削除

- トップページアクセス時に `/api/auth/sync`、`/api/user/me`、`/api/user/current-room` が呼ばれなくなる
- データベース接続が一切発生しない

### ✅ ページ読み込みの高速化

- 認証関連のJavaScript（Supabase SDK、auth-store.js等）が読み込まれない
- ページサイズの削減とパフォーマンス向上

### ✅ シンプルなUX

- 未認証・認証済み問わず、全ユーザーに同じトップページを表示
- 認証状態による表示切り替えがなく、ページフラッシュが発生しない

### ✅ サーバー負荷の軽減

- トップページアクセス時のサーバーサイド処理が最小限になる
- 認証チェックやDB問い合わせが不要

---

## 実装後の動作フロー

1. **ユーザーがトップページ（/）にアクセス**
2. **ハンドラーで静的HTMLを返す**（認証チェックなし、DB問い合わせなし）
3. **ブラウザでページを表示**（認証JSなし、SQL発行なし）
4. **ユーザーが「部屋を見る」リンクをクリック**
5. **部屋一覧ページ（/rooms）に遷移**（そこで必要に応じて認証）

---

## 修正ファイル一覧

1. `cmd/server/routes.go` - ルーティングから認証ミドルウェアを削除
2. `internal/handlers/home.go` - `StaticPage: true` を設定
3. `templates/pages/home.tmpl` - 認証JavaScript削除、部屋作成ボタン削除

---

## テスト項目

- [ ] 未認証ユーザーがトップページにアクセスできる
- [ ] 認証済みユーザーがトップページにアクセスしても `/rooms` にリダイレクトされない
- [ ] トップページで部屋作成ボタンが表示されない
- [ ] ブラウザの開発者ツールで、認証関連のAPIリクエストが発生していないことを確認
- [ ] ページのロード速度が改善されていることを確認

---

## 注意点・今後の課題

### StaticPageフラグの活用

今回導入した `StaticPage: true` フラグは、他の静的ページでも活用できます：

- `/terms` - 利用規約ページ
- `/privacy` - プライバシーポリシーページ
- `/contact` - お問い合わせページ（表示のみ）

これらのページでも認証が不要な場合、同様に `StaticPage: true` を設定することで、パフォーマンスの向上が期待できます。

### ヘッダーの表示

トップページでは `StaticPage: true` によって静的ヘッダーを描画するため、認証状態に関わらず常に以下の構成になります。

- **共通**: ロゴリンク + 「ログイン」「新規登録」ボタンのみを表示
- **認証JS不要**: Alpine.js のストア参照が無く、プロフィールメニューや部屋作成ボタンは描画されない

### 5. 認証ページでの既存セッションリダイレクト

**ファイル**: `cmd/server/routes.go`, `internal/handlers/auth.go`

**変更点**:

- `/auth/login` と `/auth/register` ルートに `withOptionalAuth` を適用し、認証ミドルウェアが `sb-access-token` クッキーを検証してユーザー情報をコンテキストに格納できるようにした。
- ログイン／新規登録ハンドラー冒頭で `middleware.GetUserFromContext` を確認し、既にセッションが有効な場合は即座に `/rooms` へリダイレクトするようにした。

**理由**: トップページのリンクを素の `<a>` のまま保ちつつ、バックエンド側でセッション判定とリダイレクトを実現することで、トップページのJavaScriptを追加せずに要件を満たすため。

### 6. 静的ページでのモバイルメニュー無効化

**ファイル**: `templates/layouts/base.tmpl`

**変更点**: `.StaticPage` のときはモバイルメニュー用コンポーネント（`$store.mobileMenu` / `$store.auth` 前提）を描画しないように条件分岐を追加。

**理由**: 認証JSを読み込まないページでは Alpine のストアが初期化されないため、対応するテンプレートを出力しないことでコンソールエラーを防止。

---

## まとめ

### 主な修正内容

1. **ルーティングから認証ミドルウェアを削除**: トップページで認証チェックが不要に
2. **StaticPageフラグの設定**: 認証関連JSの読み込みを無効化
3. **認証JavaScriptの削除**: リダイレクト処理を削除
4. **部屋作成ボタンの削除**: UIをシンプルに

### 解決した問題

- ✅ トップページアクセス時のSQL発行を完全に削除
- ✅ ページ読み込み速度の向上
- ✅ 認証状態によるページフラッシュの解消
- ✅ サーバー負荷の軽減

既存のコードへの影響を最小限に抑えつつ、トップページのパフォーマンスとUXを大幅に改善できました。
