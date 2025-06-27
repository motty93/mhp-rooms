# アーキテクチャ設計書

## 概要
MonHubは、モンスターハンターシリーズのマルチプレイヤーセッション管理を行うWebアプリケーションです。クリーンアーキテクチャの原則に基づいて設計されています。

## 技術スタック

### バックエンド
- **言語**: Go 1.22.2
- **Webフレームワーク**: Gorilla Mux
- **ORM**: GORM v2
- **認証**: Supabase Auth (gotrue-go)
- **データベース**:
  - 開発環境: PostgreSQL (Docker Compose)
  - 本番環境: Neon (Serverless PostgreSQL)

### フロントエンド
- **テンプレートエンジン**: Go標準のhtml/template
- **JavaScript**: Alpine.js v3
- **CSS**: Tailwind CSS v3
- **非同期通信**: htmx v1.9

### インフラストラクチャ
- **コンテナ**: Docker + Docker Compose
- **デプロイ**: Fly.io
- **環境変数管理**: godotenv

## ディレクトリ構造

```
mhp-rooms/
├── cmd/                           # アプリケーションエントリーポイント
│   ├── server/main.go            # Webサーバー
│   ├── migrate/main.go           # DBマイグレーション
│   └── seed/main.go              # DBシード
├── internal/                      # 内部パッケージ
│   ├── config/                    # 設定管理
│   ├── handlers/                  # HTTPハンドラー
│   │   ├── auth.go               # 認証関連
│   │   ├── rooms.go              # ルーム管理
│   │   ├── middleware.go         # ミドルウェア（未使用）
│   │   └── ...
│   ├── infrastructure/            # インフラストラクチャ層
│   │   ├── auth/                 # 認証システム
│   │   │   └── supabase/         # Supabase統合
│   │   └── persistence/          # データ永続化
│   │       └── postgres/         # PostgreSQL実装
│   ├── models/                    # データモデル
│   │   ├── user.go
│   │   ├── room.go
│   │   └── ...
│   └── repository/                # リポジトリパターン
│       ├── repository.go
│       ├── user_repository.go
│       ├── room_repository.go
│       └── ...
├── templates/                     # HTMLテンプレート
│   ├── layouts/                   # レイアウト
│   ├── pages/                     # ページテンプレート
│   └── components/                # 再利用可能コンポーネント
├── static/                        # 静的ファイル
│   ├── css/
│   ├── js/
│   └── images/
├── scripts/                       # SQLスクリプト
└── docs/                          # ドキュメント
```

## アーキテクチャレイヤー

### 1. プレゼンテーション層
- **HTTPハンドラー** (`internal/handlers/`)
  - リクエストの受信とレスポンスの返却
  - テンプレートのレンダリング
  - JSONレスポンスの生成

### 2. ビジネスロジック層
- **リポジトリ** (`internal/repository/`)
  - データアクセスロジック
  - ビジネスルールの実装
  - トランザクション管理

### 3. データモデル層
- **モデル** (`internal/models/`)
  - エンティティ定義
  - GORMタグによるマッピング
  - バリデーションルール

### 4. インフラストラクチャ層
- **認証** (`internal/infrastructure/auth/`)
  - Supabase Auth統合
  - JWT管理
- **永続化** (`internal/infrastructure/persistence/`)
  - データベース接続管理
  - マイグレーション実行

## 主要コンポーネント

### 認証システム
- Supabase Authを使用したJWT認証
- メール/パスワード認証とGoogle OAuth対応
- HttpOnlyクッキーによるセッション管理

### ルーム管理
- ルームの作成・参加・退出
- パスワード保護機能
- ゲームバージョン別の管理
- 最大4人までの人数制限

### ユーザー管理
- プロフィール機能
- ゲームバージョン別プレイヤー名
- ブロック機能

## セキュリティ

### 現在の実装
- Supabase AuthによるJWT検証
- パスワードのハッシュ化（bcrypt）
- CSRF保護（SameSiteクッキー）

### 未実装の課題
- **ミドルウェアが未使用**
  - AuthMiddleware
  - RequireAuthMiddleware
  - ProfileCompleteMiddleware
- 認証が必要なエンドポイントが保護されていない

## パフォーマンス最適化

### データベース
- 適切なインデックスの設定
- N+1問題の回避（Preload使用）
- コネクションプーリング

### フロントエンド
- Alpine.jsによる軽量なインタラクティブ機能
- htmxによる部分的なDOM更新
- 静的ファイルの効率的な配信

## デプロイメント

### 開発環境
```bash
docker-compose up -d
make migrate
make seed
make run
```

### 本番環境（Fly.io）
```bash
fly deploy
fly secrets set DATABASE_URL="..."
```

## 今後の改善点

1. **ミドルウェアの適用**
   - 認証ミドルウェアの実装
   - ルート保護の実装

2. **キャッシュレイヤー**
   - Redisの導入検討
   - セッション管理の改善

3. **モニタリング**
   - ログ収集システム
   - パフォーマンスメトリクス

## 設計原則

1. **単一責任の原則**: 各コンポーネントは1つの責任のみを持つ
2. **依存性逆転の原則**: 具体的な実装ではなく抽象に依存
3. **関心の分離**: ビジネスロジックとインフラストラクチャの分離
4. **テスタビリティ**: モックを使用した単体テストが可能な設計