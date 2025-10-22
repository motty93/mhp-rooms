# ステージング環境用 Cloud Build 設定作成 & GCS バケット構成整理

**実装時間**: 約1時間
**作成日**: 2025-10-22

## 実装概要

1. ステージング環境へのデプロイ専用の Cloud Build 設定ファイル `cloudbuild.stg.yml` を作成
2. GCS バケット名を `monhub-master` に統一
3. GCS 内のディレクトリ構成を環境最上位に統一（`{env}/種類/...`）

### 目的
- ステージング環境への安全なデプロイを実現
- ビルド済みイメージを使用してデプロイのみを実行
- 本番環境とステージング環境の明確な分離
- GCS バケット構成の一貫性向上

## 主な仕様

### 1. 環境の分離
- **サービス名**:
  - Cloud Run (main): `monhub-stg`
  - Cloud Run Jobs: `ogp-renderer-stg`
  - SSE サービス: `monhub-sse-stg`（現在はデプロイしない設定）

- **バケット設定**:
  - バケット名: `monhub-master`（本番と共有）
  - プレフィックス: `stg`（ステージング環境専用）

- **シークレット**:
  - `DATABASE_URL__stg`（ステージング環境専用のデータベース接続情報）

### 2. イメージ管理
- **本番とステージングで同じイメージを使用**
  - ステージング環境で検証したものと全く同じバイナリを本番にデプロイ可能
  - イメージタグ: `latest`（ビルド済みの最新イメージ）

### 3. デプロイ対象
- ✅ Cloud Run (main): デプロイする
- ❌ SSE サービス: デプロイしない（`_DEPLOY_SSE: "false"`）
- ✅ Cloud Run Jobs (ogp-renderer): デプロイする

## 特に注意した点

### 1. ビルドステップの削除
- ビルドは `cloudbuild.yml` で実施
- `cloudbuild.stg.yml` ではデプロイのみを実行
- 依存関係（`waitFor`）を適切に調整

### 2. イメージタグの統一
- 本番・ステージング共に `latest` タグを使用
- 環境の違いはサービス名と環境変数で管理
- シンプルで管理しやすい構成を実現

### 3. プレフィックスによる環境分離
- GCS バケットは `monhub-master` を共有
- プレフィックス（`stg` vs `prod`）で環境を分離
- コスト効率と管理の簡素化を両立

## ファイル構成

```
cloudbuild.yml       # 本番環境用（ビルド + デプロイ）
cloudbuild.stg.yml   # ステージング環境用（デプロイのみ）
```

## 今後の作業

### 必須作業
1. **Cloud Storage のセットアップ**
   - バケット `monhub-master` の作成と公開設定
   - IAM 権限の設定
   - 既存画像の移行（該当する場合）
   - **詳細は [Cloud Storage セットアップガイド](../../cloud-storage-setup.md) を参照**

2. **Secret Manager の設定**
   ```bash
   # ステージング用のデータベース接続情報を設定
   gcloud secrets create DATABASE_URL__stg \
     --replication-policy="automatic" \
     --data-file=- <<EOF
   postgresql://user:password@host/database?sslmode=require
   EOF
   ```

3. **Cloud Build トリガーの設定**
   - トリガー名: `deploy-staging`
   - ブランチ: `develop` または手動トリガー
   - 設定ファイル: `cloudbuild.stg.yml`

4. **IAM 権限の確認**
   - Cloud Build サービスアカウントに必要な権限を付与
   - Cloud Run へのデプロイ権限
   - Cloud Run Jobs へのデプロイ権限
   - Secret Manager へのアクセス権限

### 推奨作業
1. **デプロイフローの文書化**
   - ステージング環境へのデプロイ手順
   - 本番環境へのプロモーション手順

2. **モニタリングの設定**
   - ステージング環境のログ監視
   - エラーアラートの設定

3. **デプロイ戦略の検討**
   - Blue/Green デプロイの検討
   - カナリアリリースの検討

## 補足事項

### デプロイフロー（想定）
1. 開発者が `main` ブランチにマージ
2. `cloudbuild.yml` が自動実行（ビルド + 本番デプロイ）
3. 必要に応じて `cloudbuild.stg.yml` を手動実行（ステージングデプロイ）

または：
1. 開発者が `develop` ブランチにマージ
2. `cloudbuild.stg.yml` が自動実行（ステージングデプロイ）
3. 検証後、`main` ブランチにマージ
4. `cloudbuild.yml` が自動実行（ビルド + 本番デプロイ）

### 注意点
- SSE サービスは現在ステージング環境にデプロイしない設定
- 必要に応じて `_DEPLOY_SSE: "true"` に変更可能
- イメージのビルドは常に `cloudbuild.yml` で実施すること

## GCS バケット構成の整理

### 実施内容

#### 1. バケット名の統一
- 本番環境 (`cloudbuild.yml`) のバケット名を `ogp` から `monhub-master` に変更
- これにより、全環境で同一のバケット `monhub-master` を使用

#### 2. ディレクトリ構成の統一
以前は画像の種類によって構成が異なっていました：
- プロフィール画像: `{env}/avatars/...`
- 通報添付: `{env}/reports/...`
- OGP画像: `og/{env}/rooms/...` ⚠️ 不統一

これを以下のように統一：
```
monhub-master/
├── prod/
│   ├── avatars/          # プロフィール画像
│   │   └── {userID}/
│   ├── reports/          # 通報添付ファイル
│   │   └── {reportID}/
│   └── ogp/              # ← 変更
│       └── rooms/
│           └── {roomID}.png
└── stg/
    ├── avatars/
    ├── reports/
    └── ogp/              # ← 変更
        └── rooms/
```

#### 3. 変更ファイル
- `cloudbuild.yml`: バケット名を `monhub-master` に変更
- `cmd/ogp-renderer/main.go`: OGP画像のパスを変更
  - ローカル: `tmp/images/{env}/ogp/rooms/` (変更前: `tmp/images/og/{env}/rooms/`)
  - GCS: `{env}/ogp/rooms/` (変更前: `og/{env}/rooms/`)
- `templates/layouts/room_detail.tmpl`: OGP画像URLの生成パスを変更
- `Makefile`: ログメッセージのパスを変更

### メリット
1. **一貫性**: 全ての画像で同じディレクトリ構造
2. **管理性**: 環境単位での削除・バックアップが容易
3. **可読性**: パスを見れば環境と種類が一目瞭然
4. **拡張性**: 新しい画像種類を追加しやすい

### 移行について
既存のOGP画像（`og/{env}/rooms/`）は新しいパス（`{env}/ogp/rooms/`）に移動する必要があります。

```bash
# GCSでのパス移動例（gsutil使用）
gsutil -m mv gs://monhub-master/og/prod/rooms/* gs://monhub-master/prod/ogp/rooms/
gsutil -m mv gs://monhub-master/og/stg/rooms/* gs://monhub-master/stg/ogp/rooms/
```

または、既存画像は徐々に削除され、新しいパスで再生成されるのを待つこともできます。
