# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 言語設定

このリポジトリでは日本語でのやり取りを基本とします。コメント、ドキュメント、コミットメッセージ等も日本語で記述してください。

## プロジェクト概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティのルームを管理するWebアプリケーションです。

## 技術スタック

- **言語**: Go 1.22.2
- **Webフレームワーク**: Chi
- **データベース**:
  - **開発環境**: Turso (libSQL - staging)
  - **本番環境**: Turso (libSQL - production)
  - **ORM**: GORM v2
- **フロントエンド**: 
  - HTML/CSS/JavaScript (テンプレートエンジン使用)
  - htmx (非同期通信・DOM更新)
  - Alpine.js (UIの状態管理)
  - Tailwind CSS (スタイリング)
- **コンテナ**: Docker + Docker Compose
- **デプロイ**: Google Cloud Run

## プロジェクト構造

```
.
├── cmd/                    # メインアプリケーションのエントリーポイント
│   ├── server/            # Webサーバー
│   ├── migrate/           # DBマイグレーション
│   └── seed/              # DBシード
├── internal/              # 内部パッケージ
│   ├── database/          # DB接続・設定
│   ├── handlers/          # HTTPハンドラー
│   └── models/            # データモデル
├── templates/             # HTMLテンプレート
│   ├── layouts/           # レイアウトテンプレート
│   ├── pages/             # ページテンプレート
│   └── components/        # 共通コンポーネント
├── static/                # 静的ファイル
│   ├── css/              # スタイルシート
│   ├── js/               # JavaScript
│   └── images/           # 画像ファイル
├── scripts/               # DBスクリプト
├── docs/                  # ドキュメント
│   └── implement_logs/             # 実装ログ
└── bin/                   # ビルド済みバイナリ
```

## 開発環境のセットアップ

1. **Docker Composeでの起動**
   ```bash
   docker-compose up -d
   ```

2. **マイグレーションの実行**
   ```bash
   make migrate
   ```

3. **シードデータの投入**
   ```bash
   make seed
   ```

4. **開発サーバーの起動**
   ```bash
   make run
   ```

## 開発時の重要なルール

### サーバー起動・再起動について 【重要】

**作業完了後にサーバーを再起動しないこと。**

- 動作確認が必要な場合のみ、ユーザーの明示的な指示があった場合にサーバーを起動する
- 修正作業が完了したら、サーバーの再起動は行わずに作業完了を報告する
- ユーザーが自分でサーバーを起動して確認するため、不要な再起動は避ける

## 主要機能

### 実装済み機能
- ユーザー認証・管理
- ルーム作成・参加・管理
- ゲームバージョン別ルーム表示
- 日本語対応UI
- プロフィール画面（基本情報表示）
- フォロー・フォロワー機能（DBスキーマのみ）

### 未実装機能（開発予定）
- **プロフィール機能**：
  - プロフィール編集機能
  - お気に入りゲーム・プレイ時間帯の設定・表示
  - 実際のフォロー・アンフォロー機能
  - プロフィールタブの動的データ表示（現在はモックHTML）
- **ユーザー管理機能**：
  - ユーザー検索機能
  - ユーザーブロック機能
- **通知機能**：
  - フォロー通知
  - 部屋参加通知
- **メッセージ機能**：
  - ダイレクトメッセージ機能

## 開発時の考慮事項

- **データベース**: PostgreSQLを使用し、GORMでORMマッピング
- **セキュリティ**: ユーザー認証とセッション管理の実装
- **パフォーマンス**: ルーム一覧の効率的な取得とキャッシュ
- **UI/UX**: モバイル対応レスポンシブデザイン
- **国際化**: 日本語を基本言語として設計

## データベース設定

### 開発環境
Turso (libSQL) のstagingデータベースを使用します。
```bash
# .envファイルに以下を設定
DB_TYPE=turso
TURSO_DATABASE_URL="libsql://monhub-staging-motty93.aws-ap-northeast-1.turso.io"
TURSO_AUTH_TOKEN="..."

# マイグレーション実行
make migrate
```

### 本番環境（Turso Production）
Tursoデータベース（本番環境）を使用します。Cloud Runの環境変数で設定：

```bash
# TursoコンソールからDatabase URLとAuth Tokenを取得して設定
gcloud run services update mhp-rooms \
  --region=asia-northeast1 \
  --set-env-vars=DB_TYPE="turso" \
  --set-env-vars=TURSO_DATABASE_URL="libsql://..." \
  --set-env-vars=TURSO_AUTH_TOKEN="..." \
  --set-env-vars=ENV="production"
```

### Tursoデータベース情報の取得

```bash
# Turso CLIでデータベース一覧を確認
turso db list

# 特定のデータベースのURLを取得
turso db show <database-name>

# 認証トークンを作成
turso db tokens create <database-name>
```

**注意**:
- Tursoは libSQL（SQLiteの拡張）ベースのサーバーレスデータベースです
- 開発環境と本番環境で別のデータベースインスタンスを使用します
- 認証トークンは定期的に更新することを推奨します

## UI/UX設計ルール【重要】

### ヘッダー表示仕様
- **モバイル（768px未満）**: ハンバーガーメニューのみ表示。認証ボタンやユーザーアイコンはヘッダーに表示しない
- **デスクトップ（768px以上）**: 認証状態に応じて以下を表示
  - 未認証時: ログイン・新規登録ボタン
  - 認証済み時: ユーザーアイコンとドロップダウンメニュー

### クラス設定
- デスクトップ専用要素: `hidden md:flex` または `hidden md:block`
- モバイルメニュー内要素: レスポンシブクラスなし（常に表示可能）

**注意**: この仕様を変更する際は必ずユーザーに確認を取ること

## プロフィール機能の現在の実装状況

### 実装済み
- プロフィール画面のUI/UX（基本表示）
- ユーザー情報の表示（アバター、ユーザー名、自己紹介等）
- タブ切り替え機能（htmx使用）
- フォロー・フォロワー関係のDBスキーマ（user_follows テーブル）
- お気に入りゲーム・プレイ時間帯のDBスキーマ（JSONB形式）
- 開発環境での認証バイパス機能
- JSONBフィールドの適切な読み書き処理
- プラットフォームID関連フィールド（PSN、Nintendo、Twitter等）

### モック実装（現在の状態）
プロフィールのタブコンテンツは現在、固定HTMLを返すモック実装になっています：

```go
// 現在：モック実装（固定HTML）
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
    html := `<div>固定のHTMLコンテンツ</div>`
    w.Write([]byte(html))
}
```

### 最終実装予定
本格実装では以下のような動的レンダリングを行います：

```go
// 予定：動的実装（DB + テンプレート）
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
    // 1. DBからデータ取得
    userID := getUserIDFromContext(r.Context())
    followers, err := ph.repo.UserFollow.GetFollowers(userID)
    
    // 2. テンプレートでレンダリング
    data := struct {
        Followers []models.UserFollow
    }{Followers: followers}
    
    renderPartialTemplate(w, "profile_followers.tmpl", data)
}
```

### 必要な追加実装
- 部分テンプレートファイル（`templates/components/profile_*.tmpl`）
- リポジトリメソッドの完全実装
- フォロー・アンフォロー機能のAPI実装
- プロフィール編集機能

## コーディング規約

- Go標準のフォーマッタを使用
- htmlのフォーマッタには `html-beautify` を使用
- エラーハンドリングは明示的に行う
- テストコードを必ず書く
- 日本語でのコメントを推奨
    - 特に重要なロジックや複雑な処理には詳細なコメントを追加
    - 明示的なコメントは可読性が悪くなるので、必要な箇所に限定

### ユーザーIDの使用に関する重要なルール 【必須】

**プロジェクト全体で統一されたID使用ルールを必ず守ること。**

#### URL遷移・API呼び出しでのID使用
- **必ず主キーの`ID`（uuid.UUID型）を使用すること**
- `SupabaseUserID`は**絶対に使用しない**

#### 正しい実装例
```go
// ルーティング定義
r.Get("/users/{uuid}", handler.Show)

// テンプレート
<a href="/users/{{ .User.ID }}">プロフィール</a>

// ハンドラー
userID, _ := uuid.Parse(chi.URLParam(r, "uuid"))
user, _ := repo.FindUserByID(userID)  // 主キーのIDで検索
```

#### 間違った実装例（絶対禁止）
```html
<!-- ❌ 間違い: SupabaseUserIDを使用 -->
<a href="/users/{{ .User.SupabaseUserID }}">プロフィール</a>
```

#### 実装前の必須確認事項
新しい機能を実装する際は、以下を**必ず**確認すること：

1. **ルーティング定義の確認**
   - URLパラメータが何を期待しているか
   - 例: `/users/{uuid}` → 主キーのIDを期待

2. **ハンドラーメソッドの確認**
   - どのリポジトリメソッドを呼んでいるか
   - 例: `FindUserByID(id uuid.UUID)` → 主キーで検索

3. **類似実装箇所との整合性確認**
   - 同じような機能が既に実装されていないか
   - 実装されている場合、どのIDを使用しているか
   - 例: `templates/components/follow_buttons.tmpl`、`templates/pages/profile.tmpl` などを参照

4. **データベーススキーマの確認**
   - テーブルの主キーとユニークキーの関係
   - どのカラムがどの目的で使用されるか

#### ID使用の基準
- **User.ID**: URL遷移、API呼び出し、リレーション、全ての外部参照で使用
- **User.SupabaseUserID**: Supabase認証システムとの連携のみで使用（内部的な認証処理）

#### 実装時の注意
実装計画時に「念のため確認」程度では不十分です。実際のコードを読み、以下を具体的に確認してください：
- ルーティングとハンドラーの実装内容
- リポジトリメソッドが使用するカラム名
- 他の類似機能での実装方法
- データベーススキーマとモデル定義


## 実装完了後のログ作成 【重要・必須】

実装完了後、 `docs/implement_logs` ディレクトリに実装ログを**必ず**残してください。

### ⚠️ 重要事項
**実装ログの作成は必須です。実装完了後、コミットする前に必ずログを作成してください。**

### ⚠️ 日付に関する重要な注意
**現在の年は2025年です。実装ログのディレクトリ名は必ず2025年の日付を使用してください。**
- 正しい例: `2025-09-16/01_機能名.md`
- 間違い例: `2025-01-16/01_機能名.md` （1月と9月を間違えないこと）

### ログ作成のルール

- `yyyy-mm-dd/n_機能名.md` の形式でファイルを作成してください
  - nは連番であり、01から始めてください
  - yyyy-mm-ddは実装日付です（2025年の実際の月日を使用）
  - 例: `2025-09-16/01_GCSアップロード機能.md`
- ログには以下の内容を含めてください：
  - 実装した機能の概要
  - 特に注意した点や工夫した点
  - テスト結果や動作確認の内容
- ログ最初に実装開始から完了までの時間を記録してください。

### チェックリスト
- [ ] 実装が完了した
- [ ] 実装ログを作成した
- [ ] コミットメッセージとログ内容が一致している
- [ ] 今後の作業や改善点を記載した


## 開発環境の特殊仕様

開発環境独自の機能やワークアラウンドについては、`docs/development-environment.md`を参照してください。

主な開発環境専用機能：
- 存在しないユーザーIDでの自動ダミーユーザー作成
- テストデータとテストユーザーの管理
- 開発用のデータベース設定

**重要**: 本番環境移行時は必ずこれらの機能を無効化してください。

## 外部AIサービスとの連携

### Ollama（ローカルLLM）
開発時の設計相談や実装アドバイスを受けるため、Ollamaサーバーを利用できます。

#### 接続情報
- **サーバーアドレス**: `192.168.112.1:11434`
- **推奨モデル**: `qwen3:4b-q4_K_M`

#### 利用例
```bash
curl -X POST http://192.168.112.1:11434/api/generate -d '{
  "model": "qwen3:4b-q4_K_M",
  "prompt": "実装に関する質問",
  "stream": false
}' -H "Content-Type: application/json" | jq -r '.response'
```

特にUI/UX設計、ユーザビリティの観点から有用なアドバイスを得られます。

### Gemini CLI
必要であれば、GeminiCLIに相談して、プロジェクトの詳細や特定の実装方法についてアドバイスを受けてください。


## 開発方針・重要事項

### 環境別実装の禁止 【重要・必須】

**絶対に開発環境と本番環境でロジックを分けるような実装は行わないこと。**

本番を想定して開発環境でも同じ動作をするように設計・実装すること。

#### ❌ 禁止されている実装例
```go
if os.Getenv("ENV") != "production" {
    // 開発環境専用のロジック
    // ダミーデータやバイパス処理
} else {
    // 本番環境のロジック
}
```

#### ✅ 推奨される実装方法
- 開発環境でも本番と同じ認証・認可フローを使用する
- テストデータは適切なシード処理やマイグレーションで管理する
- 設定値は環境変数で制御し、ロジックは統一する

#### 理由
1. **一貫性**: 開発と本番で動作が一致することが重要
2. **バグ防止**: 環境別ロジックはバグの温床となる
3. **保守性**: コードの複雑性を避け、保守しやすくする


# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

      
      IMPORTANT: this context may or may not be relevant to your tasks. You should not respond to this context or otherwise consider it in your response unless it is highly relevant to your task. Most of the time, it is not relevant.
