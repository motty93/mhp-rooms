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
10. [放置部屋の自動削除](#放置部屋の自動削除)
11. [トラブルシューティング](#トラブルシューティング)

---

## 全体構成

```
GitHub
  ├─ staging ブランチへ push（main → staging の PR「Staging Release YYYYMMDD」をマージ）
  │    └─ Cloud Build トリガー monhub-stg-trigger (cloudbuild.stg.yml) ─→ ステージング環境
  │         ├── Docker Build & Push（App / ogp-renderer / room-cleanup）
  │         ├── Cloud Run サービスデプロイ (huntershub-stg)
  │         ├── Cloud Run Jobs デプロイ (ogp-renderer-stg)
  │         └── Cloud Run Jobs デプロイ (room-cleanup-stg)
  │
  └─ production ブランチへ push（staging → production の PR「Production Release YYYYMMDD」をマージ）
       └─ Cloud Build トリガー monhub-prd-trigger (cloudbuild.yml) ─→ 本番環境
            ├── Cloud Run サービスデプロイ (huntershub)
            ├── Cloud Run Jobs デプロイ (ogp-renderer)
            └── Cloud Run Jobs デプロイ (room-cleanup)
                ※ ステージングでビルド済みの latest イメージを使用

※ main ブランチへの push ではデプロイは走らない
```

### 使用GCPサービス

| サービス | 用途 |
|---------|------|
| Cloud Build | CI/CDパイプライン |
| Artifact Registry | Dockerイメージの保存 (`huntershub-registry`) |
| Cloud Run | Webアプリケーションのホスティング |
| Cloud Run Jobs | OGP画像生成バッチ処理、放置部屋の自動削除 |
| Cloud Scheduler | 放置部屋の自動削除 Job の定期実行 |
| Secret Manager | シークレット（DB接続情報等）の管理 |
| Cloud Storage | 静的アセット・OGP画像の保存 |

---

## 環境一覧

| 項目 | ステージング | 本番 |
|------|------------|------|
| Cloud Buildファイル | `cloudbuild.stg.yml` | `cloudbuild.yml` |
| サービス名 | `huntershub-stg` | `huntershub` |
| OGP Job名 | `ogp-renderer-stg` | `ogp-renderer` |
| 自動削除 Job名 | `room-cleanup-stg` | `room-cleanup` |
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
1. build-image (Appイメージビルド)        ─→ 3. push-image ─→ 5. deploy-main ─→ 6. deploy-sse (skip)
2. build-job-image (ogp-renderer)         ─→ 4. push-job-image ─→ 7. deploy-job
2b. build-cleanup-image (room-cleanup)    ─→ 4b. push-cleanup-image ─→ 7b. deploy-cleanup-job
```

**並列実行の最適化:**
- ステップ1・2・2bは並列でビルド
- ステップ3・4・4bはそれぞれのビルド完了後すぐにプッシュ
- ステップ5・7・7bはそれぞれのプッシュ完了後すぐにデプロイ

### 本番環境

`cloudbuild.yml` で定義。ステージングでビルド済みの `latest` イメージを使用するため、**ビルドステップがありません**。

```
1. deploy-main ─→ 2. deploy-sse (skip)
                │
                ├── 3. deploy-job (並列実行)
                └── 4. deploy-cleanup-job (並列実行)
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
| `_DEPLOY_JOB` | `true` | Cloud Run Job（OGP Renderer / room-cleanup）をデプロイするか |
| `_ROOM_INACTIVE_HOURS` | `48` | room-cleanup が「最後の活動から何時間で自動削除するか」 |

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

### Room Cleanup Job（放置部屋の自動削除）

| 設定項目 | 値 |
|---------|-----|
| タスク数 | 1 |
| 最大リトライ | 1 |
| Dockerfile | `cmd/room-cleanup/Dockerfile` |
| 環境変数 | `DB_TYPE=turso`, `ROOM_INACTIVE_HOURS=48`（`DRY_RUN=true` で対象一覧のみ表示） |

詳細は [放置部屋の自動削除](#放置部屋の自動削除) を参照。

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

### Room Cleanup (`cmd/room-cleanup/Dockerfile`)

放置部屋の自動削除専用のJobイメージです。アセットを含まない最小構成です。

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

## 放置部屋の自動削除

一定期間（既定 48 時間）活動がない募集中の部屋を自動的に解散する Cloud Run Job `room-cleanup` を、Cloud Scheduler から 1 時間ごとに実行します。

### 判定と処理内容

- **対象**: `is_active = true` の部屋のうち、しきい値以降に「作成・設定変更（`rooms.updated_at`）」「参加・退出（`room_logs`）」「チャット（`room_messages`）」のいずれもない部屋
- **処理**: ホストによる解散と同じ `DismissRoom`（メンバー全員退出・`is_active = false`）に加えて、`rooms.dismiss_reason = inactive` / `dismissed_at` を記録し、`room_logs` に `auto_dismiss`、ホストのアクティビティに「【部屋自動削除】」を残す
- **表示**: プロフィールの「作成した部屋」タブで「自動削除」ラベルと「一定期間利用がなかったため、自動的に削除されました」の注記が表示される
- 既に解散済みの部屋は対象外なので、Job を何度実行しても安全（冪等）

### 環境変数

| 変数 | 既定 | 説明 |
|------|------|------|
| `ROOM_INACTIVE_HOURS` | `48` | 最後の活動から何時間で自動削除するか（cloudbuild の `_ROOM_INACTIVE_HOURS` で設定） |
| `DRY_RUN` | `false` | `true` にすると削除せず対象一覧をログに出すだけ |
| `DB_TYPE` / `TURSO_DATABASE_URL` / `TURSO_AUTH_TOKEN` | - | 接続先 DB（Job には Secret Manager から注入） |

### 手動実行

```bash
# ローカル（.env の接続先に対して）。まず DRY_RUN で対象を確認する
make room-cleanup DRY_RUN=true
make room-cleanup

# Cloud Run Job を手動実行（確認のみ）
gcloud run jobs execute room-cleanup-stg --region=asia-northeast1 --update-env-vars DRY_RUN=true --wait

# 実行履歴とログ
gcloud run jobs executions list --job=room-cleanup-stg --region=asia-northeast1 --limit=3
```

### Cloud Scheduler の設定（環境ごとに初回のみ）

Cloud Build は Job のデプロイまでを行うため、定期実行のスケジュールは別途作成します。

```bash
PROJECT_ID=ad-hoc-rooms
REGION=asia-northeast1
JOB=room-cleanup                # ステージングは room-cleanup-stg
SA=<Job を起動するサービスアカウント>   # 例: <PROJECT_NUMBER>-compute@developer.gserviceaccount.com

# Scheduler が Job を起動できるように権限を付与
gcloud run jobs add-iam-policy-binding ${JOB} \
  --region=${REGION} \
  --member="serviceAccount:${SA}" \
  --role="roles/run.invoker"

# 毎時 0 分（JST）に実行
gcloud scheduler jobs create http ${JOB}-hourly \
  --location=${REGION} \
  --schedule="0 * * * *" \
  --time-zone="Asia/Tokyo" \
  --uri="https://run.googleapis.com/v2/projects/${PROJECT_ID}/locations/${REGION}/jobs/${JOB}:run" \
  --http-method=POST \
  --oauth-service-account-email="${SA}"

# 動作確認（即時実行）
gcloud scheduler jobs run ${JOB}-hourly --location=${REGION}
```

しきい値を変えたい場合は cloudbuild の `_ROOM_INACTIVE_HOURS` を変更して再デプロイするか、`gcloud run jobs update ${JOB} --region=${REGION} --update-env-vars ROOM_INACTIVE_HOURS=<時間>` で直接変更します。

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
