# OGP画像生成機能

## 概要

部屋情報が作成・更新されたときに、自動的にOGP（Open Graph Protocol）画像を生成する機能です。

## OGP画像の仕様

- **サイズ**: 1200x630px
- **フォーマット**: PNG
- **内容**:
  - MonHubロゴ（左上）
  - 部屋名（中央、最大2行）
  - サブタイトル「モンハンパーティ募集」
  - ゲームバージョン（右下）
  - 背景グラデーション（ゲームバージョン別配色）

## 動作モード

### 1. ローカルモード（local）

開発環境で推奨。部屋作成・更新時に自動的にローカルでOGP画像を生成します。

**設定方法**:
```bash
# .envファイル
OGP_GENERATION_MODE=local
OG_PREFIX=dev
```

**画像保存先**:
```
tmp/images/og/dev/rooms/{room_id}.png
```

### 2. Cloud Runモード（cloud）

本番環境で推奨。Cloud Run Jobsを使用してOGP画像を生成します。

**設定方法**:
```bash
# .envファイルまたは環境変数
OGP_GENERATION_MODE=cloud
PROJECT_ID=your-gcp-project-id
LOCATION=asia-northeast1
OGP_JOB_NAME=ogp-renderer
OG_BUCKET=your-ogp-bucket
OG_PREFIX=prod
```

**画像保存先**:
```
gs://{OG_BUCKET}/og/{OG_PREFIX}/rooms/{room_id}.png
```

### 3. スキップモード（skip）

OGP画像生成を完全にスキップします。

**設定方法**:
```bash
# .envファイル
OGP_GENERATION_MODE=skip
```

## 自動判定

`OGP_GENERATION_MODE`を設定しない場合、以下のように自動判定されます：

- `PROJECT_ID`が設定されている → `cloud`モード
- `PROJECT_ID`が未設定 → `local`モード

## 手動でOGP画像を生成

Makefileコマンドを使用して、特定の部屋のOGP画像を手動で生成できます。

```bash
# ROOM_IDを指定してOGP画像を生成
make generate-ogp ROOM_ID=123e4567-e89b-12d3-a456-426614174000
```

**出力例**:
```
OGP画像を生成中: ROOM_ID=123e4567-e89b-12d3-a456-426614174000
2025/10/16 15:30:45 OGP画像生成開始: room_id=123e4567-e89b-12d3-a456-426614174000, bucket=, prefix=dev
2025/10/16 15:30:45 部屋情報取得完了: name=初心者歓迎部屋, game_version=MHP3
2025/10/16 15:30:45 配色決定: game_version=MHP3
2025/10/16 15:30:45 OGP画像生成完了
2025/10/16 15:30:45 ローカル保存完了: path=tmp/images/og/dev/rooms/123e4567-e89b-12d3-a456-426614174000.png
✅ OGP画像生成完了: tmp/images/og/dev/rooms/123e4567-e89b-12d3-a456-426614174000.png
```

## 実装の流れ

### 1. 部屋作成・更新時

```go
// internal/handlers/rooms.go

// 部屋情報を更新
room.Name = req.Name
room.OGVersion++ // OGP画像バージョンをインクリメント

// OGP画像生成ジョブを非同期実行
go func() {
    ogpService := services.NewOGPJobService()
    if err := ogpService.TriggerOGPGeneration(context.Background(), room.ID); err != nil {
        log.Printf("OGP画像生成ジョブのトリガーに失敗: %v", err)
    }
}()
```

### 2. OGPジョブサービス

```go
// internal/services/ogp_job.go

func (s *OGPJobService) TriggerOGPGeneration(ctx context.Context, roomID uuid.UUID) error {
    switch s.mode {
    case "cloud":
        return s.triggerCloudRunJob(ctx, roomID)  // Cloud Run Jobs
    case "local":
        return s.generateOGPLocally(ctx, roomID)  // ローカル実行
    case "skip":
        log.Printf("OGP生成をスキップ: room_id=%s", roomID)
        return nil
    }
}
```

### 3. OGPレンダラー

```go
// cmd/ogp-renderer/main.go

func main() {
    // 環境変数から部屋ID取得
    roomID := os.Getenv("ROOM_ID")

    // データベースから部屋情報取得
    db.Preload("GameVersion").First(&room, roomID)

    // OGP画像生成
    img := generateOGPImage(&room, palette)

    // 保存（ローカル or GCS）
    if isLocalMode {
        saveToLocal(img, ogPrefix, roomID)
    } else {
        uploadToGCS(ctx, img, ogBucket, ogPrefix, roomID)
    }
}
```

## OGVersionフィールド

`rooms`テーブルの`og_version`フィールドは、OGP画像のバージョン管理に使用されます。

- 部屋作成時: `og_version = 1`
- 部屋更新時: `og_version++`（インクリメント）

**用途**:
- キャッシュバスティング: `?v={og_version}`をURLに付与して画像キャッシュを無効化
- 画像生成履歴の追跡

## トラブルシューティング

### OGP画像が生成されない

1. **環境変数の確認**:
   ```bash
   # ローカルモードの場合
   echo $OGP_GENERATION_MODE  # → local
   echo $DATABASE_URL         # → 設定されている必要がある
   ```

2. **ログの確認**:
   ```bash
   # サーバーログを確認
   tail -f logs/app.log | grep OGP
   ```

3. **手動実行でテスト**:
   ```bash
   make generate-ogp ROOM_ID=<existing-room-id>
   ```

### フォントが見つからないエラー

OGPレンダラーは日本語フォント（Noto Sans JP）が必要です。

```bash
# フォントセットアップスクリプトを実行
./scripts/setup-fonts.sh
```

### Cloud Run Jobsが実行されない

1. **環境変数の確認**:
   ```bash
   echo $PROJECT_ID
   echo $LOCATION
   echo $OGP_JOB_NAME
   ```

2. **Cloud Run Jobsのステータス確認**:
   ```bash
   gcloud run jobs describe ogp-renderer --region=asia-northeast1
   ```

3. **権限の確認**:
   - Cloud Run Jobs API が有効化されているか
   - サービスアカウントに適切な権限があるか

## 本番環境デプロイ時の注意

1. **環境変数の設定**:
   ```bash
   # Cloud Runの場合（環境変数を設定）
   # 注: PROJECT_IDとLOCATIONは同じGCPプロジェクト内では自動取得されるため不要
   gcloud run services update mhp-rooms \
     --region=asia-northeast1 \
     --set-env-vars=OGP_GENERATION_MODE=cloud \
     --set-env-vars=OGP_JOB_NAME=ogp-renderer \
     --set-env-vars=OG_BUCKET=your-bucket \
     --set-env-vars=OG_PREFIX=prod
   ```

2. **Cloud Run Jobsのデプロイ**:
   ```bash
   # Dockerイメージをビルド
   docker build -t gcr.io/your-project/ogp-renderer -f Dockerfile.ogp .

   # GCRにプッシュ
   docker push gcr.io/your-project/ogp-renderer

   # Cloud Run Jobsを作成
   gcloud run jobs create ogp-renderer \
     --image gcr.io/your-project/ogp-renderer \
     --region asia-northeast1 \
     --set-env-vars DATABASE_URL=your-db-url
   ```

3. **GCSバケットの設定**:
   - パブリックアクセスを有効化
   - CORS設定
   - キャッシュコントロールヘッダー設定

## 関連ファイル

- `internal/services/ogp_job.go` - OGPジョブサービス
- `cmd/ogp-renderer/main.go` - OGP画像レンダラー
- `internal/palette/` - ゲームバージョン別配色
- `internal/models/room.go` - Roomモデル（OGVersionフィールド）
- `internal/handlers/rooms.go` - 部屋ハンドラー（OGP生成トリガー）
- `Makefile` - `generate-ogp`コマンド
