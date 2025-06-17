# テンプレートシステム

## 概要

MHP Roomsプロジェクトでは、共通化されたテンプレートシステムを使用してUIの一貫性を保ちます。

## ディレクトリ構造

```
templates/
├── layouts/
│   └── base.html              # ベースレイアウト
├── components/
│   ├── header.html            # 共通ヘッダー
│   └── footer.html            # 共通フッター
└── pages/
    ├── home.html              # ホームページ
    ├── rooms.html             # 部屋一覧ページ
    └── ...                    # その他のページ
```

## テンプレートの構成

### 1. ベースレイアウト (`layouts/base.html`)

全ページで共通のHTMLフレームワークを定義：
- HTML構造
- メタタグ、CSS、JSの読み込み
- ヘッダー・フッター・メインコンテンツエリア

### 2. コンポーネント (`components/`)

再利用可能なUI部品：

#### ヘッダー (`header.html`)
- ロゴ・ナビゲーション
- 認証状態による表示切り替え
- レスポンシブメニュー

#### フッター (`footer.html`)
- サイト情報・リンク
- 対応ゲーム一覧

### 3. ページテンプレート (`pages/`)

各ページ固有のコンテンツ：
- `head`: ページ固有のhead内容
- `content`: メインコンテンツ

## 新しいページの作成手順

### 1. ページテンプレートを作成

```html
<!-- templates/pages/example.html -->
{{define "head"}}
<!-- ページ固有のCSS/JSがあれば記述 -->
<meta name="description" content="ページの説明">
{{end}}

{{define "content"}}
<section class="py-8">
  <div class="container mx-auto px-4">
    <h1 class="text-3xl font-bold">ページタイトル</h1>
    <!-- ページコンテンツ -->
  </div>
</section>
{{end}}
```

### 2. ハンドラーを追加

```go
// internal/handlers/handlers.go
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
    data := TemplateData{
        Title:   "ページタイトル",
        HasHero: false, // ヒーローセクションの有無
        PageData: struct{
            // ページ固有のデータ
        }{
            // データの値
        },
    }
    renderTemplate(w, "example.html", data)
}
```

### 3. ルートを設定

```go
// cmd/server/main.go
r.HandleFunc("/example", handlers.ExampleHandler).Methods("GET")
```

## TemplateData 構造体

```go
type TemplateData struct {
    Title    string      // ページタイトル
    HasHero  bool        // ヒーローセクションの有無
    User     interface{} // ユーザー情報
    PageData interface{} // ページ固有のデータ
}
```

### フィールドの説明

- **Title**: ブラウザのタイトルバーに表示される
- **HasHero**: `true`の場合、メインエリアにpadding-topが適用されない
- **User**: 認証済みユーザーの情報（将来実装）
- **PageData**: ページごとに異なるデータを格納

## ページタイプ別のベストプラクティス

### ヒーローセクションありのページ（トップページ等）
```go
data := TemplateData{
    Title:   "ホーム",
    HasHero: true,  // ヒーローセクションでpadding調整
}
```

### 通常ページ（一覧、詳細等）
```go
data := TemplateData{
    Title:   "部屋一覧",
    HasHero: false, // 通常のpadding-top適用
}
```

## スタイリングガイドライン

### レスポンシブデザイン
- `container mx-auto px-4`: 基本コンテナ
- `grid md:grid-cols-2 lg:grid-cols-3`: レスポンシブグリッド
- `hidden md:flex`: レスポンシブ表示切り替え

### 色彩パレット
- プライマリ: `bg-gray-800`, `text-gray-800`
- セカンダリ: `bg-gray-100`, `text-gray-600`
- アクセント: `bg-green-100`, `bg-yellow-100`, `bg-red-100`

### 間隔
- セクション間: `py-8`, `py-16`
- 要素間: `mb-4`, `mb-6`, `mb-8`
- 内部余白: `px-4`, `px-6`, `py-2`, `py-3`

## 認証状態の表示制御

ヘッダーコンポーネントでは以下のIDで認証状態を制御：

- `#auth-buttons`: 未認証時のボタン
- `#user-menu`: 認証済み時のユーザーメニュー
- `#nav-menu`: 認証済み時のナビゲーション

JavaScript（将来実装）で表示切り替えを行います。

## パフォーマンス考慮事項

- テンプレートはリクエストごとにパースされる（開発時）
- 本番環境では事前コンパイル・キャッシュを検討
- 静的アセット（CSS/JS）のキャッシュ設定

## トラブルシューティング

### よくあるエラー

1. **Template parsing error**
   - ファイルパスの確認
   - テンプレート構文の確認

2. **Template execution error**
   - `{{define}}` と `{{template}}` の名前一致確認
   - データ構造の確認

3. **スタイルが適用されない**
   - CSS ファイルのパス確認
   - MIMEタイプの設定確認

### デバッグ方法

1. テンプレートファイルの存在確認
2. ブラウザの開発者ツールでネットワークエラー確認
3. サーバーログでエラー詳細確認