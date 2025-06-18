# Docker開発環境セットアップ

## 概要

このドキュメントでは、MHP RoomsプロジェクトをローカルのDocker環境で実行する方法を説明します。

## 前提条件

以下がインストールされていることを確認してください：

- [Docker](https://docs.docker.com/get-docker/) (20.10.0+)
- [Docker Compose](https://docs.docker.com/compose/install/) (2.0.0+)

## サービス構成

Docker Composeで以下のサービスが起動します：

| サービス名 | 説明               | ポート | URL                                                             |
| ---------- | ------------------ | ------ | --------------------------------------------------------------- |
| app        | Goアプリケーション | 8080   | http://localhost:8080                                           |
| db         | PostgreSQL 15      | 5432   | postgresql://mhp_user:mhp_password@localhost:5432/mhp_rooms_dev |
| redis      | Redis 7            | 6379   | redis://localhost:6379                                          |
| pgadmin    | pgAdmin 4          | 8081   | http://localhost:8081                                           |

## 環境構築手順

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd mhp-rooms
```

### 2. 環境ファイルの作成（オプション）

```bash
# .env.local ファイルを作成（必要に応じて設定をカスタマイズ）
cp .env.example .env.local
```

### 3. Docker Composeでサービス起動

```bash
# バックグラウンドで全サービスを起動
docker-compose up -d

# ログを確認
docker-compose logs -f app
```

### 4. アプリケーションにアクセス

- **Webアプリ**: http://localhost:8080
- **pgAdmin**: http://localhost:8081
  - Email: admin@mhp-rooms.local
  - Password: admin

## 開発時のコマンド

### サービスの操作

```bash
# 全サービス起動
docker-compose up -d

# 特定のサービスのみ起動
docker-compose up -d db redis

# サービス停止
docker-compose down

# ボリュームも含めて完全削除
docker-compose down -v

# ログ確認
docker-compose logs -f app
docker-compose logs -f db
```

### アプリケーションの再ビルド

```bash
# アプリケーションのDockerイメージを再ビルド
docker-compose build app

# キャッシュを使わずに再ビルド
docker-compose build --no-cache app

# 再ビルド後に起動
docker-compose up -d --build app
```

### データベース操作

```bash
# PostgreSQLコンテナに接続
docker-compose exec db psql -U mhp_user -d mhp_rooms_dev

# データベースのリセット
docker-compose down -v
docker-compose up -d db
```

### Redisの操作

```bash
# Redisコンテナに接続
docker-compose exec redis redis-cli

# Redisデータのクリア
docker-compose exec redis redis-cli FLUSHALL
```

## pgAdmin設定

pgAdminでPostgreSQLに接続する設定：

1. http://localhost:8081 にアクセス
2. 以下の情報でログイン：
   - Email: `admin@mhp-rooms.local`
   - Password: `admin`
3. 新しいサーバーを追加：
   - **Name**: MHP Rooms Local
   - **Host**: `db`
   - **Port**: `5432`
   - **Database**: `mhp_rooms_dev`
   - **Username**: `mhp_user`
   - **Password**: `mhp_password`

## 環境変数

アプリケーションで使用される主な環境変数：

| 変数名       | 説明                     | デフォルト値                                                           |
| ------------ | ------------------------ | ---------------------------------------------------------------------- |
| PORT         | アプリケーションのポート | 8080                                                                   |
| DATABASE_URL | PostgreSQL接続文字列     | postgres://mhp_user:mhp_password@db:5432/mhp_rooms_dev?sslmode=disable |
| REDIS_URL    | Redis接続文字列          | redis://redis:6379                                                     |

## トラブルシューティング

### ポートが既に使用されている

```bash
# ポートの使用状況を確認
lsof -i :8080
lsof -i :5432

# プロセスを停止してから再度実行
```

### データベース接続エラー

```bash
# データベースの健全性確認
docker-compose exec db pg_isready -U mhp_user -d mhp_rooms_dev

# データベースログを確認
docker-compose logs db
```

### ボリュームの問題

```bash
# 全ボリュームを削除して再作成
docker-compose down -v
docker volume prune
docker-compose up -d
```

### イメージの問題

```bash
# 全てのイメージを再ビルド
docker-compose build --no-cache
docker-compose up -d
```

## 本番環境との違い

| 項目           | 開発環境            | 本番環境          |
| -------------- | ------------------- | ----------------- |
| データベース   | PostgreSQL (Docker) | Fly.io PostgreSQL |
| Redis          | Redis (Docker)      | Fly.io Redis      |
| SSL            | 無効                | 有効              |
| ログレベル     | DEBUG               | INFO              |
| ホットリロード | 有効                | 無効              |

## 次のステップ

1. **マイグレーション実装**: GORMを使用したデータベーススキーマの自動生成
2. **認証機能**: Supabase Authとの連携実装
3. **WebSocket**: リアルタイム通信機能の実装
4. **テスト環境**: テスト用のDocker Compose設定追加
