# Cloud Run での マイグレーション実行方法

本番環境（Cloud Run / Cloud Run Jobs）でのデータベースマイグレーション実行方法について説明します。

## 推奨方法：Cloud Run Jobs を使用

### 1. Cloud Run Jobs でマイグレーション実行

```bash
# Jobを作成して実行
gcloud run jobs create migrate-job \
  --image=gcr.io/YOUR_PROJECT/mhp-rooms \
  --set-env-vars="DB_TYPE=turso" \
  --region=asia-northeast1 \
  --task-count=1 \
  --max-retries=0 \
  --command="./migrate"

# Job実行
gcloud run jobs execute migrate-job --region=asia-northeast1
```

### 2. 既存イメージを使用した簡単な実行

```bash
gcloud run jobs create migrate-turso \
  --image=YOUR_EXISTING_IMAGE \
  --set-env-vars="DB_TYPE=turso" \
  --region=asia-northeast1 \
  --command="go" \
  --args="run,cmd/migrate/main.go"
```

## 代替方法

### マイグレーション専用イメージの作成

```dockerfile
# Dockerfile.migrate
FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o migrate cmd/migrate/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/migrate .
CMD ["./migrate"]
```

### 既存のCloud Runサービスで一回限り実行

```bash
# 既存サービスでマイグレーションコマンドを実行
gcloud run services update YOUR_SERVICE_NAME \
  --region=asia-northeast1 \
  --set-env-vars="DB_TYPE=turso,RUN_MIGRATION=true"
```

## 環境変数の設定

### 必要な環境変数

- `DB_TYPE=turso` : Tursoデータベースを使用することを指定
- その他のデータベース接続情報（`DATABASE_URL`等）はCloud Runのシークレットとして設定済み

### シークレットの確認

```bash
# 現在のシークレット確認
gcloud run services describe YOUR_SERVICE_NAME --region=asia-northeast1
```

## 注意事項

1. **Cloud Run Jobs**が一回限りのマイグレーション実行には最適
2. マイグレーションは冪等性があることを確認してから実行
3. 本番データベースへの影響を考慮し、必要に応じてバックアップを事前に取得
4. マイグレーション実行前にテスト環境での動作確認を推奨

## ローカル環境でのテスト

本番実行前に、ローカル環境でturso設定でのマイグレーションをテスト：

```bash
DB_TYPE=turso go run cmd/migrate/main.go
```