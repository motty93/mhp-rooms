# デプロイガイド

HuntersHubのデプロイに関する包括的なドキュメントです。

## 目次

1. [全体構成](#全体構成)
2. [環境一覧](#環境一覧)
3. [デプロイフロー](#デプロイフロー)
4. [Cloud Buildの設定](#cloud-buildの設定)
5. [Cloud Runサービス構成](#cloud-runサービス構成)
6. [シークレット管理](#シークレット管理)
7. [Dockerイメージ](#dockerイメージ)
8. [手動デプロイ](#手動デプロイ)
9. [マイグレーション](#マイグレーション)
10. [トラブルシューティング](#トラブルシューティング)

---

## 全体構成

```
GitHub (main)
  │
  ├─ Cloud Build (cloudbuild.stg.yml) ─→ ステージング環境
  │    ├── Docker Build & Push
  │    ├── Cloud Run サービスデプロイ (huntershub-stg)
  │    └── Cloud Run Jobs デプロイ (ogp-renderer-stg)
  │
  └─ Cloud Build (cloudbuild.yml) ─→ 本番環境
       ├── Cloud Run サービスデプロイ (huntershub)
       └── Cloud Run Jobs デプロイ (ogp-renderer)
           ※ ステージングでビルド済みの latest イメージを使用
```

### 使用GCPサービス

| サービス | 用途 |
|---------|------|
| Cloud Build | CI/CDパイプライン |
| Artifact Registry | Dockerイメージの保存 (`huntershub-registry`) |
| Cloud Run | Webアプリケーションのホスティング |
| Cloud Run Jobs | OGP画像生成バッチ処理 |
| Secret Manager | シークレット（DB接続情報等）の管理 |
| Cloud Storage | 静的アセット・OGP画像の保存 |

---

## 環境一覧

| 項目 | ステージング | 本番 |
|------|------------|------|
| Cloud Buildファイル | `cloudbuild.stg.yml` | `cloudbuild.yml` |
| サービス名 | `huntershub-stg` | `huntershub` |
| OGP Job名 | `ogp-renderer-stg` | `ogp-renderer` |
| リージョン | `asia-northeast1` | `asia-northeast1` |
| ENV | `staging` | `production` |
| DB | Turso (staging) | Turso (production) |
| アセットプレフィックス | `stg` | `prod` |
| OGPプレフィックス | `stg` | `prod` |
| GA4 | 有効（ステージング用ID） | 有効（本番用ID: `G-5T7T5SCL50`） |

---

## デプロイフロー

### ステージング環境

`cloudbuild.stg.yml` で定義。ビルドからデプロイまで全工程を実行します。

```
1. build-image (Appイメージビルド)  ─┐─→ 3. push-image ─→ 5. deploy-main ─→ 6. deploy-sse (skip)
                                     │
2. build-job-image (Jobイメージビルド)┘─→ 4. push-job-image ─→ 7. deploy-job
```

**並列実行の最適化:**
- ステップ1と2は並列でビルド
- ステップ3と4はそれぞれのビルド完了後すぐにプッシュ
- ステップ5と7はそれぞれのプッシュ完了後すぐにデプロイ

### 本番環境

`cloudbuild.yml` で定義。ステージングでビルド済みの `latest` イメージを使用するため、**ビルドステップがありません**。

```
1. deploy-main ─→ 2. deploy-sse (skip)
                │
                └── 3. deploy-job (並列実行)
```

**重要:** 本番デプロイは必ずステージングでの検証後に実行してください。本番はステージングと同じイメージ（`latest`タグ）を使用します。

---

## Cloud Buildの設定

### 設定ファイル

| ファイル | 環境 | ビルド有無 | マシンタイプ |
|---------|------|----------|------------|
| `cloudbuild.stg.yml` | ステージング | あり | `E2_MEDIUM` |
| `cloudbuild.yml` | 本番 | なし（latestイメージ使用） | `UNSPECIFIED` |

### Artifact Registry

- リポジトリ名: `huntershub-registry`
- リージョン: `asia-northeast1`
- イメージ:
  - `asia-northeast1-docker.pkg.dev/<PROJECT_ID>/huntershub-registry/huntershub-app`
  - `asia-northeast1-docker.pkg.dev/<PROJECT_ID>/huntershub-registry/ogp-renderer`

### タグ戦略

- ステージング: `$SHORT_SHA`（gitコミットハッシュ）+ `latest`
- 本番: `latest`（ステージングでプッシュ済み）

### デプロイフラグ

`cloudbuild.yml` / `cloudbuild.stg.yml` の `substitutions` で制御:

| 変数 | デフォルト | 説明 |
|------|----------|------|
| `_DEPLOY_SSE` | `false` | SSEサービスをデプロイするか |
| `_DEPLOY_JOB` | `true` | OGP Renderer Jobをデプロイするか |

---

## Cloud Runサービス構成

### メインアプリケーション

| 設定項目 | 値 |
|---------|-----|
| メモリ | 512Mi |
| CPU | 1 vCPU |
| ポート | 8080 |
| タイムアウト | 3600秒 |
| 最小インスタンス | 0 |
| 最大インスタンス | 10 |
| 同時実行数 | 10 |
| 認証 | 未認証許可 (`--allow-unauthenticated`) |
| Ingress | `all` |

### SSEサービス（現在無効）

| 設定項目 | 値 |
|---------|-----|
| メモリ | 256Mi |
| CPU | 1 vCPU |
| 最大インスタンス | 5 |
| 同時実行数 | 100 |
| SERVICE_MODE | `sse` |

### OGP Renderer Job

| 設定項目 | 値 |
|---------|-----|
| タスク数 | 1 |
| 最大リトライ | 2 |
| Dockerfile | `cmd/ogp-renderer/Dockerfile` |

### 環境変数

Cloud Runに設定される環境変数:

| 変数 | 説明 |
|------|------|
| `ENV` | 環境名 (`staging` / `production`) |
| `DB_TYPE` | データベース種別 (`turso`) |
| `SERVICE_MODE` | サービスモード (`main` / `sse`) |
| `RUN_MIGRATION` | 起動時マイグレーション実行 (`true`) |
| `SUPABASE_URL` | Supabase URL |
| `SUPABASE_ANON_KEY` | Supabase Anon Key |
| `GCS_BUCKET` | GCSバケット名 |
| `GCS_PRIVATE_BUCKET` | GCSプライベートバケット名 |
| `BASE_PUBLIC_ASSET_URL` | 公開アセットのベースURL |
| `ASSET_PREFIX` | アセットのプレフィックス (`stg` / `prod`) |
| `OG_BUCKET` | OGP画像保存先バケット |
| `OG_PREFIX` | OGP画像のプレフィックス (`stg` / `prod`) |
| `OGP_JOB_NAME` | OGP生成Jobの名前 |
| `OGP_GENERATION_MODE` | OGP生成モード (`cloud`) |
| `GA_ENABLED` | GA4計測の有効化 (`true` / `false`) |
| `GA_MEASUREMENT_ID` | GA4測定ID |

---

## シークレット管理

Secret Managerで管理され、Cloud Runに `--set-secrets` で注入されます。

| シークレット名 (stg) | シークレット名 (prod) | 用途 |
|---------------------|---------------------|------|
| `TURSO_DATABASE_URL__stg` | `TURSO_DATABASE_URL__prod` | TursoデータベースURL |
| `TURSO_AUTH_TOKEN__stg` | `TURSO_AUTH_TOKEN__prod` | Turso認証トークン |
| `SUPABASE_JWT_SECRET__stg` | `SUPABASE_JWT_SECRET__prod` | JWT検証シークレット |
| `DISCORD_WEBHOOK_URL__stg` | `DISCORD_WEBHOOK_URL__prod` | Discord通知Webhook |

詳細な設定手順は [Secret Manager セットアップガイド](./secret-manager-setup.md) を参照してください。

---

## Dockerイメージ

### メインアプリケーション (`Dockerfile`)

マルチステージビルド:

1. **ビルドステージ** (`golang:1.24-alpine`)
   - 依存関係のダウンロード
   - 静的ファイル生成 (`cmd/generate_info` — 更新情報・ロードマップ)
   - CGO有効でGoバイナリをビルド
2. **ランタイムステージ** (`alpine:3.18`)
   - 非rootユーザー (`appuser:1001`) で実行
   - ヘルスチェック: `curl -f http://localhost:8080/health`
   - ポート: 8080

### OGP Renderer (`cmd/ogp-renderer/Dockerfile`)

OGP画像生成専用のJobイメージです。

---

## 手動デプロイ

通常はCloud Buildトリガーで自動デプロイされますが、手動で実行する場合:

### ステージングへのデプロイ

```bash
# Cloud Buildを手動実行
gcloud builds submit \
  --config=cloudbuild.stg.yml \
  --region=asia-northeast1
```

### 本番へのデプロイ

```bash
# Cloud Buildを手動実行
gcloud builds submit \
  --config=cloudbuild.yml \
  --region=asia-northeast1 \
  --no-source
```

### 特定のサービスのみ再デプロイ

```bash
# メインサービスのみ再デプロイ（イメージ変更なし、環境変数更新など）
gcloud run deploy huntershub \
  --image asia-northeast1-docker.pkg.dev/<PROJECT_ID>/huntershub-registry/huntershub-app:latest \
  --region asia-northeast1

# ステージング
gcloud run deploy huntershub-stg \
  --image asia-northeast1-docker.pkg.dev/<PROJECT_ID>/huntershub-registry/huntershub-app:latest \
  --region asia-northeast1
```

---

## マイグレーション

### 自動マイグレーション

Cloud Runサービスに `RUN_MIGRATION=true` が設定されているため、**デプロイ時に自動的にマイグレーションが実行**されます。

### 手動マイグレーション

ローカルから手動で実行する場合:

```bash
# ローカル開発環境
make migrate-dev

# ビルド済みバイナリで実行
make migrate
```

Cloud Run Jobsで実行する場合は [Cloud Run マイグレーション実行方法](./cloud-run-migration.md) を参照してください。

---

## トラブルシューティング

### デプロイが失敗する

```bash
# Cloud Buildのログを確認
gcloud builds list --limit=5 --region=asia-northeast1
gcloud builds log <BUILD_ID> --region=asia-northeast1

# Cloud Runサービスのログを確認
gcloud run services logs read huntershub --region=asia-northeast1 --limit=50
gcloud run services logs read huntershub-stg --region=asia-northeast1 --limit=50
```

### コンテナが起動しない

```bash
# Cloud Runのリビジョン状態を確認
gcloud run revisions list --service=huntershub --region=asia-northeast1

# 最新リビジョンの詳細を確認
gcloud run revisions describe <REVISION_NAME> --region=asia-northeast1
```

### シークレットにアクセスできない

```bash
# シークレットの存在確認
gcloud secrets list | grep huntershub

# IAM権限の確認
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

gcloud secrets get-iam-policy TURSO_DATABASE_URL__prod \
  --filter="bindings.members:serviceAccount:${SERVICE_ACCOUNT}"
```

### ヘルスチェックが失敗する

```bash
# ローカルでヘルスチェックを確認
curl -f http://localhost:8080/health
```

ヘルスチェックは30秒間隔・3秒タイムアウト・起動猶予5秒・3回リトライで設定されています。

---

## 関連ドキュメント

- [インフラ構成](./infra.md)
- [アーキテクチャ設計書](./architecture.md)
- [Secret Manager セットアップガイド](./secret-manager-setup.md)
- [Cloud Storage セットアップガイド](./cloud-storage-setup.md)
- [Cloud Run マイグレーション実行方法](./cloud-run-migration.md)
- [OGP画像生成](./ogp-image-generation.md)

---

**最終更新日**: 2026-03-12
