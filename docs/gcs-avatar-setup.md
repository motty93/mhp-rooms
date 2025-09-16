# Google Cloud Storage アバター画像アップロード設定ガイド

## 概要

プロフィール編集画面でのアバター画像アップロード機能が実装されました。Google Cloud Storageを使用してallUsersに公開されたバケットに環境別フォルダ（dev/prod）で画像を保存します。

## 必要な環境変数

### 必須設定

```bash
# GCSバケット名（allUsers設定済み）
GCS_BUCKET=your-avatar-bucket-name

# 公開URL（CDN使用時は差し替え）
BASE_PUBLIC_ASSET_URL=https://storage.googleapis.com/your-avatar-bucket-name

# 環境別プレフィックス（dev / stg / prod）
ASSET_PREFIX=dev  # 開発環境: dev, 本番環境: prod
```

### オプション設定（デフォルト値あり）

```bash
# アップロード制限設定
MAX_UPLOAD_BYTES=10485760  # デフォルト: 10MB
ALLOW_CONTENT_TYPES=image/jpeg,image/png,image/webp  # デフォルト
```

## GCSバケット設定

1. **バケットの作成**
   ```bash
   gsutil mb gs://your-avatar-bucket-name
   ```

2. **allUsersに公開設定**
   ```bash
   gsutil iam ch allUsers:objectViewer gs://your-avatar-bucket-name
   ```

3. **CORSポリシー設定**（必要に応じて）
   ```json
   [{
     "origin": ["https://your-domain.com"],
     "method": ["GET", "POST"],
     "responseHeader": ["Content-Type"],
     "maxAgeSeconds": 3600
   }]
   ```

## フォルダ構造

```
your-bucket/
├── dev/
│   └── avatars/
│       └── {user-id}/
│           └── avatar-{hash}.jpg
└── prod/
    └── avatars/
        └── {user-id}/
            └── avatar-{hash}.jpg
```

## 実装詳細

### ファイル保存形式
- **パス**: `{ASSET_PREFIX}/avatars/{user-id}/{basename}-{hash12}.{ext}`
- **例**: `dev/avatars/123e4567-e89b-12d3-a456-426614174000/avatar-abc123def456.jpg`

### セキュリティ
- ファイルサイズ制限: 10MB
- 許可形式: JPEG, PNG, WebP
- ファイル名サニタイズ
- 重複防止のためのハッシュ付き命名

### キャッシュ設定
- **Cache-Control**: `public, max-age=31536000, immutable`
- ファイル名にハッシュが含まれるため、更新時は新しいファイル名で保存

## CDN導入時の設定変更

CDNを導入する場合は、`BASE_PUBLIC_ASSET_URL`のみを変更：

```bash
# 例: Cloudflare使用時
BASE_PUBLIC_ASSET_URL=https://img.your-domain.com
```

## トラブルシューティング

1. **GCSクライアントの初期化エラー**
   - Google Application Default Credentials（ADC）を設定
   - サービスアカウントキーを設定

2. **アップロードエラー**
   - バケットのアクセス権限を確認
   - ファイルサイズ制限を確認

3. **画像が表示されない**
   - allUsersの公開設定を確認
   - BASE_PUBLIC_ASSET_URLが正しいか確認

## 使用方法

1. プロフィール編集画面でアバター画像をクリック
2. ファイルを選択（JPEG/PNG/WebP、10MB以下）
3. 自動でアップロードされ、プロフィール表示に戻る
4. 画像はGCSに保存され、DBにURLが更新される