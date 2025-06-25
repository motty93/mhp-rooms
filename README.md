# MHP Rooms - モンスターハンターポータブル アドホックパーティ ルーム管理システム

## 概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティのルームを管理するWebアプリケーションです。

### 主な機能

- ✅ ユーザー登録・ログイン機能
- ✅ ルーム作成・参加・管理
- ✅ リアルタイムルーム状態更新
- ✅ ゲームバージョン別ルーム表示
- ✅ パスワード付きルーム作成
- ✅ ユーザーブロック機能
- ✅ 日本語対応UI

## 技術スタック

### バックエンド
- **言語**: Go 1.22.2
- **Webフレームワーク**: Gorilla Mux
- **ORM**: GORM v2
- **データベース**: PostgreSQL

### フロントエンド
- **テンプレートエンジン**: Go HTML/Template
- **UIライブラリ**: htmx (非同期通信・DOM更新)
- **状態管理**: Alpine.js (インライン記述)
- **スタイリング**: Tailwind CSS (CDN)

### インフラ・ツール
- **コンテナ**: Docker & Docker Compose
- **デプロイ**: Fly.io
- **ビルドツール**: Make

## プロジェクト構造

```
.
├── cmd/                    # アプリケーションエントリーポイント
│   ├── server/            # Webサーバー
│   ├── migrate/           # DBマイグレーション
│   └── seed/              # シードデータ投入
├── internal/              # 内部パッケージ
│   ├── config/            # アプリケーション設定
│   ├── database/          # DB接続・設定
│   ├── handlers/          # HTTPハンドラー
│   ├── models/            # データモデル
│   ├── repository/        # データアクセス層
│   └── utils/             # ユーティリティ関数
├── templates/             # HTMLテンプレート
│   ├── layouts/           # レイアウトテンプレート
│   ├── pages/             # ページテンプレート
│   └── components/        # 共通コンポーネント
├── static/                # 静的ファイル
│   ├── css/              # スタイルシート
│   ├── js/               # JavaScript
│   │   └── vendor/       # サードパーティライブラリ
│   └── images/           # 画像ファイル
│       └── icons/        # アイコン画像
├── scripts/               # DBスクリプト
├── docs/                  # ドキュメント
│   └── logs/             # 実装ログ
├── bin/                   # ビルド済みバイナリ
├── Makefile              # ビルドタスク
├── compose.yml           # Docker Compose設定
├── fly.toml              # Fly.io設定
├── CLAUDE.md             # AI開発支援ガイド
└── README.md
```

## 開発環境のセットアップ

### 前提条件

- Docker & Docker Compose
- Go 1.22.2以上（ローカル開発時）
- Make

### セットアップ手順

1. **リポジトリのクローン**
   ```bash
   git clone https://github.com/motty93/mhp-rooms.git
   cd mhp-rooms
   ```

2. **Dockerコンテナの起動**
   ```bash
   docker compose up -d
   ```

3. **データベースマイグレーション**
   ```bash
   make migrate
   ```

4. **シードデータの投入**（開発用データ）
   ```bash
   make seed
   ```

5. **開発サーバーの起動**
   ```bash
   make run
   ```

アプリケーションは http://localhost:8080 でアクセス可能です。

## 利用可能なコマンド

```bash
make build         # アプリケーションをビルド
make run           # アプリケーションを実行
make dev           # 開発サーバーを起動（ホットリロード付き）
make test          # テストを実行
make lint          # リンターを実行
make fmt           # コードをフォーマット
make clean         # ビルド成果物をクリーンアップ
make migrate       # DBマイグレーションを実行
make seed          # シードデータを投入
make container-up  # Dockerコンテナを起動
make container-down # Dockerコンテナを停止
```

## 環境変数

アプリケーションは以下の環境変数を使用します：

### データベース設定
- `DATABASE_URL`: PostgreSQL接続文字列（優先使用）
- `DB_HOST`: データベースホスト
- `DB_USER`: データベースユーザー名
- `DB_PASSWORD`: データベースパスワード
- `DB_NAME`: データベース名
- `DB_PORT`: データベースポート（デフォルト: 5432）
- `DB_SSLMODE`: SSL接続モード（デフォルト: disable）

### アプリケーション設定
- `PORT`: サーバーポート（デフォルト: 8080）
- `ENV`: 実行環境（development/production）
- `SESSION_SECRET`: セッションシークレットキー
- `JWT_SECRET`: JWT署名用シークレットキー

### Google OAuth設定
- `GOOGLE_CLIENT_ID`: Google OAuth クライアントID
- `GOOGLE_CLIENT_SECRET`: Google OAuth クライアントシークレット
- `GOOGLE_REDIRECT_URL`: Google OAuth リダイレクトURL

## API仕様

詳細なAPI仕様は [docs/api-design.md](docs/api-design.md) を参照してください。

## アーキテクチャ

システムアーキテクチャの詳細は [docs/architecture.md](docs/architecture.md) を参照してください。

## データベーススキーマ

データベース設計の詳細は [docs/db-schema.md](docs/db-schema.md) を参照してください。

## ライセンス

MIT

## AI開発支援

[CLAUDE.md](./CLAUDE.md) にAI開発支援のための指示が含まれています。AIを使用してコードの生成や改善を行う場合は、このファイルを参照してください。

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを作成して変更内容について議論してください。

## お問い合わせ

質問や提案がある場合は、[Issues](https://github.com/motty93/mhp-rooms/issues) でお知らせください。
