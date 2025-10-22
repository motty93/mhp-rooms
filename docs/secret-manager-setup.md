# Secret Manager セットアップガイド

このドキュメントは、MonHub プロジェクトで使用する Secret Manager のシークレット設定をまとめたものです。

## 必要なシークレット一覧

### ステージング環境

| シークレット名 | 用途 | 取得方法 |
|--------------|------|---------|
| `TURSO_DATABASE_URL__stg` | TursoデータベースURL | Turso CLI: `turso db show <db-name>` |
| `TURSO_AUTH_TOKEN__stg` | Turso認証トークン | Turso CLI: `turso db tokens create <db-name>` |
| `SUPABASE_JWT_SECRET__stg` | Supabase JWT検証シークレット | Supabaseダッシュボード → Settings → API → JWT Secret |
| `DISCORD_WEBHOOK_URL__stg` | Discord通知用WebhookURL | Discordサーバー設定 → 連携サービス → ウェブフック |

### 本番環境

| シークレット名 | 用途 | 取得方法 |
|--------------|------|---------|
| `TURSO_DATABASE_URL__prod` | TursoデータベースURL | Turso CLI: `turso db show <db-name>` |
| `TURSO_AUTH_TOKEN__prod` | Turso認証トークン | Turso CLI: `turso db tokens create <db-name>` |
| `SUPABASE_JWT_SECRET__prod` | Supabase JWT検証シークレット | Supabaseダッシュボード → Settings → API → JWT Secret |
| `DISCORD_WEBHOOK_URL__prod` | Discord通知用WebhookURL | Discordサーバー設定 → 連携サービス → ウェブフック |

---

## シークレットの作成方法

### 1. Turso Database URL

```bash
# ステージング環境
gcloud secrets create TURSO_DATABASE_URL__stg \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
libsql://your-staging-db.turso.io
EOF

# 本番環境
gcloud secrets create TURSO_DATABASE_URL__prod \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
libsql://your-prod-db.turso.io
EOF
```

**Turso Database URLの取得方法**:
```bash
# データベース一覧を確認
turso db list

# 特定のデータベースの詳細を確認
turso db show <database-name>
```

### 2. Turso Auth Token

```bash
# ステージング環境
gcloud secrets create TURSO_AUTH_TOKEN__stg \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
your-staging-token-here
EOF

# 本番環境
gcloud secrets create TURSO_AUTH_TOKEN__prod \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
your-prod-token-here
EOF
```

**Turso Auth Tokenの取得方法**:
```bash
# 新しいトークンを作成
turso db tokens create <database-name>

# または、既存のトークンを確認（注意: トークンは一度しか表示されません）
turso db tokens list <database-name>
```

### 3. Supabase JWT Secret

```bash
# ステージング環境
gcloud secrets create SUPABASE_JWT_SECRET__stg \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
your-jwt-secret-here
EOF

# 本番環境
gcloud secrets create SUPABASE_JWT_SECRET__prod \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
your-jwt-secret-here
EOF
```

**Supabase JWT Secretの取得方法**:
1. Supabaseダッシュボードにログイン
2. プロジェクトを選択
3. **Settings** → **API**
4. **JWT Secret** をコピー（⚠️ "service_role" keyではありません！）

### 4. Discord Webhook URL

```bash
# ステージング環境
gcloud secrets create DISCORD_WEBHOOK_URL__stg \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
EOF

# 本番環境
gcloud secrets create DISCORD_WEBHOOK_URL__prod \
  --replication-policy="automatic" \
  --data-file=- <<'EOF'
https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
EOF
```

**Discord Webhook URLの取得方法**:
1. Discordサーバーの設定を開く
2. **連携サービス** → **ウェブフック**
3. **新しいウェブフック** をクリック
4. ウェブフックURLをコピー

---

## シークレットの確認

### シークレット一覧の確認

```bash
gcloud secrets list
```

### 特定のシークレットの存在確認

```bash
# ステージング環境
gcloud secrets describe TURSO_DATABASE_URL__stg
gcloud secrets describe TURSO_AUTH_TOKEN__stg
gcloud secrets describe SUPABASE_JWT_SECRET__stg
gcloud secrets describe DISCORD_WEBHOOK_URL__stg

# 本番環境
gcloud secrets describe TURSO_DATABASE_URL__prod
gcloud secrets describe TURSO_AUTH_TOKEN__prod
gcloud secrets describe SUPABASE_JWT_SECRET__prod
gcloud secrets describe DISCORD_WEBHOOK_URL__prod
```

### シークレットの値を確認（開発環境のみ推奨）

```bash
# ⚠️ 注意: 本番環境では実行しないこと
gcloud secrets versions access latest --secret="TURSO_DATABASE_URL__stg"
```

---

## IAM 権限の設定

Cloud Run と Cloud Run Jobs からシークレットにアクセスできるように権限を設定します。

```bash
# プロジェクト情報の取得
PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")

# Cloud Run サービスアカウント（デフォルト）
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# すべてのシークレットに権限を付与
for SECRET in TURSO_DATABASE_URL__stg TURSO_AUTH_TOKEN__stg SUPABASE_JWT_SECRET__stg DISCORD_WEBHOOK_URL__stg TURSO_DATABASE_URL__prod TURSO_AUTH_TOKEN__prod SUPABASE_JWT_SECRET__prod DISCORD_WEBHOOK_URL__prod
do
  echo "権限を付与中: $SECRET"
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/secretmanager.secretAccessor"
done
```

---

## シークレットの更新

既存のシークレットの値を更新する場合：

```bash
# 新しいバージョンを追加
echo -n "new-secret-value" | gcloud secrets versions add TURSO_DATABASE_URL__stg --data-file=-

# または、ファイルから
gcloud secrets versions add TURSO_DATABASE_URL__stg --data-file=/tmp/new_value.txt
```

---

## トラブルシューティング

### 問題1: Cloud Runからシークレットにアクセスできない

**症状**: デプロイ時に「Secret not found」エラー

**解決策**:
```bash
# シークレットが存在するか確認
gcloud secrets list | grep TURSO_DATABASE_URL__stg

# 存在しない場合は作成
gcloud secrets create TURSO_DATABASE_URL__stg \
  --replication-policy="automatic" \
  --data-file=-
```

### 問題2: 権限エラー

**症状**: 「Permission denied」エラー

**解決策**:
```bash
# サービスアカウントに権限を付与
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")
SERVICE_ACCOUNT="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

gcloud secrets add-iam-policy-binding TURSO_DATABASE_URL__stg \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"
```

### 問題3: シークレットの値が間違っている

**症状**: アプリケーションが起動しない、データベース接続エラー

**解決策**:
```bash
# シークレットの値を確認
gcloud secrets versions access latest --secret="TURSO_DATABASE_URL__stg"

# 間違っている場合は新しいバージョンを追加
echo -n "correct-value" | gcloud secrets versions add TURSO_DATABASE_URL__stg --data-file=-

# Cloud Runを再デプロイ（新しいバージョンを読み込むため）
```

---

## チェックリスト

セットアップが完了したら、以下のチェックリストで確認してください：

### ステージング環境
- [ ] `TURSO_DATABASE_URL__stg` が作成されている
- [ ] `TURSO_AUTH_TOKEN__stg` が作成されている
- [ ] `SUPABASE_JWT_SECRET__stg` が作成されている
- [ ] `DISCORD_WEBHOOK_URL__stg` が作成されている
- [ ] すべてのシークレットにIAM権限が付与されている

### 本番環境
- [ ] `TURSO_DATABASE_URL__prod` が作成されている
- [ ] `TURSO_AUTH_TOKEN__prod` が作成されている
- [ ] `SUPABASE_JWT_SECRET__prod` が作成されている
- [ ] `DISCORD_WEBHOOK_URL__prod` が作成されている
- [ ] すべてのシークレットにIAM権限が付与されている

---

## 関連ドキュメント

- [cloudbuild.yml](../cloudbuild.yml) - 本番環境のビルド設定
- [cloudbuild.stg.yml](../cloudbuild.stg.yml) - ステージング環境のデプロイ設定
- [Cloud Storage セットアップガイド](./cloud-storage-setup.md)
- [Google Secret Manager ドキュメント](https://cloud.google.com/secret-manager/docs)

---

**最終更新日**: 2025-10-22
