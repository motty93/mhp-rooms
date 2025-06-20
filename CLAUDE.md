# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 言語設定

このリポジトリでは日本語でのやり取りを基本とします。コメント、ドキュメント、コミットメッセージ等も日本語で記述してください。

## プロジェクト概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティのルームを管理するWebアプリケーションです。

## 技術スタック

- **言語**: Go 1.22.2
- **Webフレームワーク**: Gorilla Mux
- **データベース**: PostgreSQL (GORM v2使用)
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
- リアルタイムルーム状態更新
- ゲームバージョン別ルーム表示
- 日本語対応UI

## 開発時の考慮事項

- **データベース**: PostgreSQLを使用し、GORMでORMマッピング
- **セキュリティ**: ユーザー認証とセッション管理の実装
- **パフォーマンス**: ルーム一覧の効率的な取得とキャッシュ
- **UI/UX**: モバイル対応レスポンシブデザイン
- **国際化**: 日本語を基本言語として設計

## AI開発注意事項

実装完了後、要件定義ディレクトリ `docs/logs` に実装ログを**必ず**残してください。<br/>
`yyyy-mm-dd/n_機能名.md` の形式でファイルを作成してください。nは連番です。<br/>

## コーディング規約

- Go標準のフォーマッタを使用
- htmlのフォーマッタには `html-beautify` を使用
- エラーハンドリングは明示的に行う
- テストコードを必ず書く
- 日本語でのコメントを推奨
    - 特に重要なロジックや複雑な処理には詳細なコメントを追加
    - 明示的なコメントは可読性が悪くなるので、必要な箇所に限定
