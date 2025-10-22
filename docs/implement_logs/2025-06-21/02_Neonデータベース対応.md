# Neonデータベース対応

## 実装日時
2025-06-20

## 目的
本番環境でNeon（Serverless PostgreSQL）を使用するため、DATABASE_URL環境変数への対応を追加。

## 実装内容

### 1. データベース接続の修正
`internal/database/database.go`の`getDSN()`関数を修正：

#### 変更点
- `DATABASE_URL`環境変数を優先的に使用
- 既存の個別環境変数はフォールバックとして保持
- 本番環境では自動的にSSL接続を要求

#### 修正後の動作
1. `DATABASE_URL`が設定されている場合は優先使用
2. 設定されていない場合は既存の個別環境変数から構築
3. 環境（`ENV`）に応じてSSLモードを自動設定
   - `ENV=production`: `sslmode=require`
   - `ENV=development`: `sslmode=disable`

### 2. ドキュメントの更新
`CLAUDE.md`に以下を追加：

#### 技術スタック
- 開発環境: PostgreSQL (Docker Compose)
- 本番環境: Neon (Serverless PostgreSQL)

#### データベース設定セクション
- 開発環境の設定手順
- 本番環境（Neon）の設定方法（2つの方法）
- 環境変数の優先順位の説明

## 設定方法

### 最も簡単な方法（推奨）
```bash
gcloud run services update mhp-rooms \
  --region=asia-northeast1 \
  --set-env-vars=DATABASE_URL="postgresql://username:password@ep-xxx.region.neon.tech/database?sslmode=require" \
  --set-env-vars=ENV="production"
```

### 従来の方法（フォールバック）
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

## 利点

1. **簡単な設定**: NeonのConnection Stringをそのまま使用可能
2. **後方互換性**: 既存の個別環境変数設定も継続使用可能
3. **自動SSL設定**: 本番環境では自動的にSSL接続
4. **開発との分離**: 開発環境は従来通りDocker Compose使用

## 修正ファイル
- `internal/database/database.go`
- `CLAUDE.md`

## テスト方法
1. 開発環境: `make migrate` で動作確認
2. 本番環境: DATABASE_URL設定後にCloud Runへデプロイで確認