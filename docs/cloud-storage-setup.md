# Cloud Storage セットアップガイド

このドキュメントは、MonHub プロジェクトで使用する Google Cloud Storage (GCS) バケットのセットアップ手順をまとめたものです。

## 目次

1. [概要](#概要)
2. [バケット構成](#バケット構成)
3. [セットアップ手順](#セットアップ手順)
4. [確認コマンド](#確認コマンド)
5. [トラブルシューティング](#トラブルシューティング)

---

## 概要

MonHub では以下の3つのバケットを使用します：

| バケット名 | 用途 | 公開設定 | 環境変数 |
|-----------|------|---------|---------|
| `monhub-master` | OGP画像 | 公開 | `OG_BUCKET` |
| 任意の名前 | プロフィール画像 | 公開 | `GCS_BUCKET` |
| 任意の名前 | 通報添付ファイル | プライベート | `GCS_PRIVATE_BUCKET` |

### ディレクトリ構成

```
monhub-master/
├── prod/
│   ├── avatars/          # プロフィール画像
│   │   └── {userID}/
│   │       └── avatar-{hash}.jpg
│   ├── reports/          # 通報添付ファイル（プライベート）
│   │   └── {reportID}/
│   │       └── attachment-{hash}.png
│   └── ogp/              # OGP画像
│       └── rooms/
│           └── {roomID}.png
└── stg/
    ├── avatars/
    ├── reports/
    └── ogp/
        └── rooms/
```

---

## バケット構成

### 1. OGP画像用バケット（`monhub-master`）

- **用途**: 部屋詳細ページのOGP画像
- **公開設定**: 公開読み取り可能
- **リージョン**: `asia-northeast1`
- **ストレージクラス**: Standard

### 2. プロフィール画像用バケット

- **用途**: ユーザーのプロフィール画像（アバター）
- **公開設定**: 公開読み取り可能
- **環境変数**: `GCS_BUCKET`, `BASE_PUBLIC_ASSET_URL`

### 3. 通報添付ファイル用バケット

- **用途**: ユーザー通報時の添付画像
- **公開設定**: プライベート（署名付きURLでアクセス）
- **環境変数**: `GCS_PRIVATE_BUCKET`

---

## セットアップ手順

### 前提条件

- Google Cloud SDK (`gcloud`) がインストール済み
- 適切なプロジェクトに認証済み
- 必要な権限（Storage Admin）を保持

```bash
# プロジェクトの確認
gcloud config get-value project

# 認証の確認
gcloud auth list
```

---

### ステップ 1: バケットの作成

#### 1-1. OGP画像用バケット

```bash
# バケットの存在確認
gsutil ls gs://monhub-master 2>/dev/null && echo "バケットは既に存在します" || echo "バケットを作成する必要があります"

# バケットの作成（存在しない場合）
gsutil mb -l asia-northeast1 gs://monhub-master

# 作成確認
gsutil ls -L -b gs://monhub-master
```

#### 1-2. プロフィール画像用バケット（オプション）

OGP画像と同じバケット（`monhub-master`）を使用する場合はスキップ可能です。

```bash
# 別バケットを使用する場合
gsutil mb -l asia-northeast1 gs://monhub-avatars
```

#### 1-3. 通報添付ファイル用バケット（プライベート）

```bash
# プライベートバケットの作成
gsutil mb -l asia-northeast1 gs://monhub-reports-private
```

---

### ステップ 2: 公開設定

OGP画像とプロフィール画像は公開アクセスが必要です。

```bash
# monhub-master を公開読み取り可能に設定
gsutil iam ch allUsers:objectViewer gs://monhub-master

# 確認
gsutil iam get gs://monhub-master | grep allUsers
```

**⚠️ 注意**: 通報用のプライベートバケットは公開しないでください！

---

### ステップ 3: CORS 設定（現在は不要）

**現在のアーキテクチャでは CORS 設定は不要です。**

理由：
- 画像アップロードはCloud Runサーバー経由で実施
- 画像表示は `<img>` タグで行われ、CORSの制限を受けない

#### CORSが必要になるケース

将来的に以下の機能を実装する場合のみ、CORS設定が必要になります：

- ブラウザから直接GCSにアップロード（署名付きURL使用）
- JavaScriptの `fetch()` で画像データを取得して処理
- Canvas API や WebGL で画像を操作

#### CORS設定の例（将来必要になった場合）

<details>
<summary>CORS設定手順を表示</summary>

**cors.json:**
```json
[
  {
    "origin": [
      "https://your-production-domain.com",
      "https://your-staging-domain.com",
      "http://localhost:8080"
    ],
    "method": ["GET", "HEAD", "PUT", "POST"],
    "responseHeader": ["Content-Type", "Cache-Control"],
    "maxAgeSeconds": 3600
  }
]
```

**適用コマンド:**
```bash
gsutil cors set cors.json gs://monhub-master
gsutil cors get gs://monhub-master
```

</details>

---

### ステップ 4: 既存画像の移行

既存のOGP画像がある場合、新しいディレクトリ構成に移行します。

#### 4-1. 現在の構成確認

```bash
# 古いパス（og/{env}/rooms/）の画像を確認
gsutil ls -r gs://monhub-master/og/
```

#### 4-2. パスの移行

```bash
# 本番環境の画像を移行
gsutil -m mv gs://monhub-master/og/prod/rooms/* gs://monhub-master/prod/ogp/rooms/

# ステージング環境の画像を移行
gsutil -m mv gs://monhub-master/og/stg/rooms/* gs://monhub-master/stg/ogp/rooms/

# 空のディレクトリを削除（オプション）
gsutil rm -r gs://monhub-master/og/
```

**⚠️ 注意**:
- 移行中はOGP画像が一時的にアクセスできなくなります
- ダウンタイムを避けたい場合は、コピー（`cp`）してから削除（`rm`）してください

```bash
# より安全な移行方法（コピー → 確認 → 削除）
gsutil -m cp -r gs://monhub-master/og/prod/rooms/* gs://monhub-master/prod/ogp/rooms/
gsutil -m cp -r gs://monhub-master/og/stg/rooms/* gs://monhub-master/stg/ogp/rooms/

# 確認後、古いパスを削除
gsutil -m rm -r gs://monhub-master/og/
```

---

### ステップ 5: IAM 権限の設定

Cloud Run と Cloud Run Jobs からバケットにアクセスできるように権限を設定します。

#### 5-1. Cloud Build サービスアカウント

```bash
# プロジェクト番号を取得
PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")

# Cloud Build サービスアカウント
CLOUDBUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

# Storage Object Admin 権限を付与
gsutil iam ch serviceAccount:${CLOUDBUILD_SA}:objectAdmin gs://monhub-master

# 確認
gsutil iam get gs://monhub-master | grep cloudbuild
```

#### 5-2. Cloud Run サービスアカウント

```bash
# デフォルトの Compute Engine サービスアカウント
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# 権限を付与
gsutil iam ch serviceAccount:${COMPUTE_SA}:objectAdmin gs://monhub-master

# 通報用プライベートバケットにも権限を付与
gsutil iam ch serviceAccount:${COMPUTE_SA}:objectAdmin gs://monhub-reports-private
```

---

### ステップ 6: ライフサイクルポリシー（推奨）

古い画像を自動削除してストレージコストを削減します。

#### 6-1. ライフサイクルポリシーファイルの作成

```json
// lifecycle.json
{
  "lifecycle": {
    "rule": [
      {
        "action": {
          "type": "Delete"
        },
        "condition": {
          "age": 90,
          "matchesPrefix": ["prod/ogp/rooms/", "stg/ogp/rooms/"]
        },
        "description": "90日以上経過したOGP画像を削除"
      }
    ]
  }
}
```

#### 6-2. ポリシーの適用

```bash
# ライフサイクルポリシーを設定
gsutil lifecycle set lifecycle.json gs://monhub-master

# 確認
gsutil lifecycle get gs://monhub-master
```

**💡 ヒント**: OGP画像は部屋が更新されたり削除されたりすると再生成されるため、古い画像は自動削除しても問題ありません。

---

### ステップ 7: 環境変数の設定

Cloud Run に必要な環境変数を設定します。

#### 7-1. 本番環境

```bash
gcloud run services update monhub \
  --region=asia-northeast1 \
  --set-env-vars="OG_BUCKET=monhub-master,OG_PREFIX=prod,GCS_BUCKET=monhub-master,BASE_PUBLIC_ASSET_URL=https://storage.googleapis.com/monhub-master,GCS_PRIVATE_BUCKET=monhub-reports-private,ASSET_PREFIX=prod"
```

#### 7-2. ステージング環境

```bash
gcloud run services update monhub-stg \
  --region=asia-northeast1 \
  --set-env-vars="OG_BUCKET=monhub-master,OG_PREFIX=stg,GCS_BUCKET=monhub-master,BASE_PUBLIC_ASSET_URL=https://storage.googleapis.com/monhub-master,GCS_PRIVATE_BUCKET=monhub-reports-private,ASSET_PREFIX=stg"
```

---

## 確認コマンド

セットアップ後、以下のコマンドで設定を確認してください。

### バケット一覧

```bash
gsutil ls
```

### バケット内のファイル確認

```bash
# monhub-master の中身
gsutil ls -r gs://monhub-master/

# 本番環境のOGP画像
gsutil ls gs://monhub-master/prod/ogp/rooms/

# ステージング環境のOGP画像
gsutil ls gs://monhub-master/stg/ogp/rooms/
```

### IAM設定の確認

```bash
# バケットのIAM設定
gsutil iam get gs://monhub-master

# 特定のサービスアカウントの権限確認
gsutil iam get gs://monhub-master | grep -A5 "cloudbuild"
```

### ライフサイクルポリシーの確認

```bash
gsutil lifecycle get gs://monhub-master
```

### 公開URLのテスト

```bash
# 例: OGP画像にアクセス可能か確認
curl -I https://storage.googleapis.com/monhub-master/prod/ogp/rooms/YOUR_ROOM_ID.png
```

---

## トラブルシューティング

### 問題 1: 画像がアクセスできない

**症状**: ブラウザで画像URLにアクセスすると403エラー

**解決策**:
```bash
# 公開設定を確認
gsutil iam get gs://monhub-master | grep allUsers

# 公開設定がない場合は追加
gsutil iam ch allUsers:objectViewer gs://monhub-master
```

### 問題 2: Cloud Run からアップロードできない

**症状**: OGP画像の生成に失敗する

**解決策**:
```bash
# サービスアカウントの権限を確認
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) --format="value(projectNumber)")
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

# 権限を付与
gsutil iam ch serviceAccount:${COMPUTE_SA}:objectAdmin gs://monhub-master
```

### 問題 3: 古いパスの画像が残っている

**症状**: `og/prod/rooms/` に画像が残っている

**解決策**:
```bash
# 古いパスの画像を削除
gsutil -m rm -r gs://monhub-master/og/
```

---

## 関連ドキュメント

- [Google Cloud Storage ドキュメント](https://cloud.google.com/storage/docs)
- [cloudbuild.yml](../cloudbuild.yml) - 本番環境のビルド設定
- [cloudbuild.stg.yml](../cloudbuild.stg.yml) - ステージング環境のデプロイ設定
- [実装ログ: ステージング環境用Cloud Build設定](./implement_logs/2025-10-22/02_ステージング環境用Cloud%20Build設定.md)

---

## チェックリスト

セットアップが完了したら、以下のチェックリストで確認してください：

- [ ] `monhub-master` バケットが作成されている
- [ ] バケットが公開読み取り可能になっている
- [ ] Cloud Build サービスアカウントに権限が付与されている
- [ ] Cloud Run サービスアカウントに権限が付与されている
- [ ] 既存の画像が新しいパスに移行されている（該当する場合）
- [ ] ライフサイクルポリシーが設定されている（推奨）
- [ ] Cloud Run の環境変数が正しく設定されている
- [ ] 画像URLが公開アクセス可能であることを確認した

---

**最終更新日**: 2025-10-22
