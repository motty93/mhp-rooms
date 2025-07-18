# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 言語設定

このリポジトリでは日本語でのやり取りを基本とします。コメント、ドキュメント、コミットメッセージ等も日本語で記述してください。

## プロジェクト概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティのルームを管理するWebアプリケーションです。

## 技術スタック

- **言語**: Go 1.22.2
- **Webフレームワーク**: Gorilla Mux
- **データベース**: 
  - **開発環境**: PostgreSQL (Docker Compose)
  - **本番環境**: Neon (Serverless PostgreSQL)
  - **ORM**: GORM v2
- **フロントエンド**: 
  - HTML/CSS/JavaScript (テンプレートエンジン使用)
  - htmx (非同期通信・DOM更新)
  - Alpine.js (UIの状態管理)
  - Tailwind CSS (スタイリング)
- **コンテナ**: Docker + Docker Compose
- **デプロイ**: Fly.io

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
│   └── logs/             # 実装ログ
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

## 主要機能

- ユーザー認証・管理
- ルーム作成・参加・管理
- ゲームバージョン別ルーム表示
- 日本語対応UI

## 開発時の考慮事項

- **データベース**: PostgreSQLを使用し、GORMでORMマッピング
- **セキュリティ**: ユーザー認証とセッション管理の実装
- **パフォーマンス**: ルーム一覧の効率的な取得とキャッシュ
- **UI/UX**: モバイル対応レスポンシブデザイン
- **国際化**: 日本語を基本言語として設計

## データベース設定

### 開発環境
Docker Composeで起動するPostgreSQLを使用します。
```bash
make container-up  # DBとRedisを起動
make migrate       # マイグレーション実行
```

### 本番環境（Neon）
Neonデータベースを使用します。以下の2つの方法で設定できます：

#### 方法1: DATABASE_URL（推奨・最も簡単）
```bash
# NeonコンソールからConnection Stringをコピーして設定
fly secrets set DATABASE_URL="postgresql://username:password@ep-xxx.region.neon.tech/database?sslmode=require"
fly secrets set ENV="production"
```

#### 方法2: 個別の環境変数
```bash
fly secrets set DB_HOST="ep-xxx.region.neon.tech"
fly secrets set DB_USER="your-username"
fly secrets set DB_PASSWORD="your-password"
fly secrets set DB_NAME="your-database"
fly secrets set DB_SSLMODE="require"
fly secrets set ENV="production"
```

**注意**: DATABASE_URLが設定されている場合は、個別の環境変数より優先して使用されます。

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

## コーディング規約

- Go標準のフォーマッタを使用
- htmlのフォーマッタには `html-beautify` を使用
- エラーハンドリングは明示的に行う
- テストコードを必ず書く
- 日本語でのコメントを推奨
    - 特に重要なロジックや複雑な処理には詳細なコメントを追加
    - 明示的なコメントは可読性が悪くなるので、必要な箇所に限定


## 実装完了後のログ作成 【重要・必須】

実装完了後、 `docs/logs` ディレクトリに実装ログを**必ず**残してください。

### ⚠️ 重要事項
**実装ログの作成は必須です。実装完了後、コミットする前に必ずログを作成してください。**

### ログ作成のルール

- `yyyy-mm-dd/n_機能名.md` の形式でファイルを作成してください
  - nは連番であり、01から始めてください
  - yyyy-mm-ddは実装日付です
  - 例: `2025-06-21/01_ルーム作成機能.md`
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


## その他
必要であれば、GeminiCLIに相談して、プロジェクトの詳細や特定の実装方法についてアドバイスを受けてください。

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

      
      IMPORTANT: this context may or may not be relevant to your tasks. You should not respond to this context or otherwise consider it in your response unless it is highly relevant to your task. Most of the time, it is not relevant.