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
  - **開発環境**: PostgreSQL (Docker Compose)
  - **本番環境**: Neon (Serverless PostgreSQL)
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
gcloud run services update mhp-rooms \
  --region=asia-northeast1 \
  --set-env-vars=DATABASE_URL="postgresql://username:password@ep-xxx.region.neon.tech/database?sslmode=require" \
  --set-env-vars=ENV="production"
```

#### 方法2: 個別の環境変数
```bash
gcloud run services update mhp-rooms \
  --region=asia-northeast1 \
  --set-env-vars=DB_HOST="ep-xxx.region.neon.tech" \
  --set-env-vars=DB_USER="your-username" \
  --set-env-vars=DB_PASSWORD="your-password" \
  --set-env-vars=DB_NAME="your-database" \
  --set-env-vars=DB_SSLMODE="require" \
  --set-env-vars=ENV="production"
```

**注意**: DATABASE_URLが設定されている場合は、個別の環境変数より優先して使用されます。


## コーディング規約

- Go標準のフォーマッタを使用
- htmlのフォーマッタには `html-beautify` を使用
- エラーハンドリングは明示的に行う
- テストコードを必ず書く
- 日本語でのコメントを推奨
    - 特に重要なロジックや複雑な処理には詳細なコメントを追加
    - 明示的なコメントは可読性が悪くなるので、必要な箇所に限定


## 実装完了後のログ作成

実装完了後、 `docs/implement_logs` ディレクトリに実装ログを**必ず**残してください。

- `yyyy-mm-dd/n_機能名.md` の形式でファイルを作成してください
  - nは連番であり、01から始めてください
  - yyyy-mm-ddは実装日付です
  - 例: `2025-06-21/01_ルーム作成機能.md`
- ログには以下の内容を含めてください：
  - 実装した機能の概要
  - 特に注意した点や工夫した点
  - テスト結果や動作確認の内容
- ログ最初に実装開始から完了までの時間を記録してください。


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
