# OGP画像生成機能の実装

**実装日**: 2025年10月15日
**Issue**: #53
**実装時間**: 約4時間

## 概要

部屋詳細ページのシェア時に表示されるOGP（Open Graph Protocol）画像を自動生成する機能を実装しました。Cloud Run Jobsを使用して、部屋作成・更新時にゲームバージョンに応じた配色のOGP画像（1200×630 PNG）を生成し、GCSにアップロードします。

## 実装内容

### 1. DBスキーマ更新

- **ファイル**: `scripts/add_og_version_to_rooms.sql`, `internal/models/room.go`
- **内容**: `rooms`テーブルに`og_version`カラムを追加（INT、デフォルト0）
- **目的**: OGP画像のキャッシュバスティング用バージョン番号

### 2. 配色パレット定義

- **ファイル**: `internal/palette/game_version.go`
- **内容**:
  - ゲームバージョン別の配色定義（MHP, MHP2, MHP2G, MHP3）
  - グラデーション用の上下カラー、アクセントカラー
- **特徴**: 各ゲームに対応した視覚的な差別化

### 3. OGP画像生成Job

- **ファイル**: `cmd/ogp-renderer/main.go`, `cmd/ogp-renderer/Dockerfile`
- **機能**:
  - 環境変数（ROOM_ID, OG_BUCKET, DATABASE_URL）から部屋情報を取得
  - `fogleman/gg`ライブラリで画像生成
  - ゲームバージョンに応じた背景グラデーション
  - 角丸矩形フレーム、ゲームアイコン、部屋名（最大2行）、サブ情報の描画
  - GCSへアップロード（Cache-Control: public, max-age=31536000, immutable）
- **レイアウト**:
  - サイズ: 1200×630 PNG
  - アイコン: 左上 (60,60) 128px
  - タイトル: 左寄せ (60,260) NotoSansJP Bold 64px
  - サブ情報: (60,400) NotoSansJP Regular 32px

### 4. Cloud Run Jobs実行サービス

- **ファイル**: `internal/services/ogp_job.go`
- **機能**:
  - Webアプリケーションから非同期にCloud Run Jobsを実行
  - 環境変数で実行を制御（未設定時はスキップ）
  - エラーが発生してもメイン処理に影響を与えない

### 5. ハンドラー修正

- **ファイル**: `internal/handlers/rooms.go`
- **内容**:
  - `CreateRoom`: 部屋作成時に`og_version=1`を設定し、OGPジョブを実行
  - `UpdateRoom`: 部屋更新時に`og_version`をインクリメントし、OGPジョブを実行
- **実装方法**: goroutineで非同期実行し、失敗してもメイン処理を妨げない

### 6. OGPメタタグ追加

- **ファイル**: `templates/layouts/room_detail.tmpl`, `internal/helpers/template.go`
- **内容**:
  - og:image, og:title, og:description等のメタタグ追加
  - Twitter Card対応
  - 環境変数からバケット名とプレフィックスを取得する`getEnv`関数追加

### 7. フォント準備

- **ファイル**: `cmd/ogp-renderer/assets/fonts/README.md`
- **内容**:
  - NotoSansJP（Bold/Regular）の配置場所とダウンロード方法を記載
  - Google Fontsから取得可能（OFLライセンス）

### 8. Cloud Build設定

- **ファイル**: `cloudbuild.yml`（既存）
- **内容**: OGP生成Job用のイメージビルドとデプロイ設定が既に含まれている

## 技術的ポイント

### 1. 非同期処理

部屋作成・更新のレスポンス速度を維持するため、OGP画像生成は非同期で実行：

```go
go func() {
    ogpService := services.NewOGPJobService()
    if err := ogpService.TriggerOGPGeneration(context.Background(), room.ID); err != nil {
        log.Printf("OGP画像生成ジョブのトリガーに失敗: %v", err)
    }
}()
```

### 2. キャッシュバスティング

`og_version`を部屋作成・更新時にインクリメントし、URLのクエリパラメータとして付与：

```html
<meta property="og:image" content="https://storage.googleapis.com/{bucket}/og/{env}/rooms/{id}.png?v={og_version}">
```

### 3. 環境変数による制御

開発・ステージング・本番環境で異なるバケットやプレフィックスを使用可能：

- `OG_BUCKET`: GCSバケット名
- `OG_PREFIX`: 環境プレフィックス（dev/stg/prod）
- `PROJECT_ID`, `LOCATION`: GCP設定
- `OGP_JOB_NAME`: Cloud Run Jobs名

## 残タスク

### 必須

1. **フォントファイルのダウンロード**:
   ```bash
   cd cmd/ogp-renderer/assets/fonts
   curl -L "https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Bold.ttf" -o NotoSansJP-Bold.ttf
   curl -L "https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Regular.ttf" -o NotoSansJP-Regular.ttf
   ```

2. **DBマイグレーション実行**:
   ```bash
   psql $DATABASE_URL < scripts/add_og_version_to_rooms.sql
   ```

3. **環境変数設定**（本番環境）:
   ```bash
   fly secrets set OGP_JOB_NAME="ogp-renderer"
   fly secrets set OG_BUCKET="myapp-og-images"
   fly secrets set OG_PREFIX="prod"
   fly secrets set SITE_URL="https://monhub.app"
   ```

4. **GCSバケットの作成と権限設定**:
   - バケット作成: `gsutil mb gs://myapp-og-images`
   - 公開読み取り権限付与: `gsutil iam ch allUsers:objectViewer gs://myapp-og-images`

5. **Cloud Run Jobsのデプロイ**:
   Cloud Buildトリガーを実行するか、手動でデプロイ

### 推奨

1. **OGP画像のプレビュー機能**: 管理画面やデバッグ用に画像プレビュー機能を追加
2. **エラーハンドリングの強化**: Job実行失敗時の通知やリトライ機能
3. **画像最適化**: PNG圧縮やWebP対応

## 動作確認

1. 部屋を新規作成し、`og_version`が1に設定されることを確認
2. 部屋を更新し、`og_version`がインクリメントされることを確認
3. Cloud Run Jobsが実行され、GCSに画像がアップロードされることを確認
4. 部屋詳細ページのHTMLソースでOGPメタタグが正しく出力されることを確認
5. SNS（Twitter/Facebook等）でシェアし、OGP画像が表示されることを確認

## 参考資料

- [Open Graph Protocol](https://ogp.me/)
- [Twitter Cards](https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/abouts-cards)
- [fogleman/gg Documentation](https://github.com/fogleman/gg)
- [Google Cloud Storage](https://cloud.google.com/storage/docs)
- [Cloud Run Jobs](https://cloud.google.com/run/docs/create-jobs)

## 注意事項

- OGP画像生成は非同期処理のため、部屋作成直後は画像が存在しない可能性があります
- GCSバケットは公開読み取り可能にする必要があります（OGP表示のため）
- フォントファイルはリポジトリに含めず、デプロイ時にダウンロードまたは手動配置してください
- 開発環境でもCloud Run Jobsを実行する場合は、環境変数の設定が必要です
