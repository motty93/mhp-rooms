# Cloud Run 推奨スペック（Go + htmx SSR + SSE構成）

## リージョン
- `asia-northeast1`（東京）

## サービス分割
- **web**（通常SSR用）
  - ページレンダリングや通常のAPI処理
- **sse**（SSE専用）
  - イベントストリームのみを担当

## リソース設定
- **メモリ/CPU**：512MiB / 1 vCPU
- **タイムアウト**：3600秒（SSE長時間接続対応）
- **最小インスタンス**：0（必要に応じて1へ）
- **最大インスタンス**
  - web: 10
  - sse: 50

## 同時実行（Concurrency）
- **web**：8（SSR用途の初期値として推奨）
- **sse**：1（長時間接続安定化のため）

## ネットワーク/Ingress
- **Ingress**：`all`（インターネット公開）
- **固定アウトバウンドIP**：NeonDBや外部DBでIP制限を使う場合は  
  - Direct VPC egress  
  - または VPC Connector + Cloud NAT（静的IP）を構成

## 認証
- Supabase Auth でアプリ側制御
- Cloud Run 側は「未認証呼び出し許可」

## 環境変数（.env対応）

### アプリケーション設定
- `PORT=8080`
- `ENV=development`

### データベース設定
- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=mhp_user`
- `DB_PASSWORD=mhp_password`
- `DB_NAME=mhp_rooms_dev`
- `DB_SSLMODE=disable`
- `DATABASE_URL="postgres://mhp_user:mhp_password@localhost:5432/mhp_rooms_dev?sslmode=disable"`

### ログ設定
- `LOG_LEVEL=debug`
- `DEBUG_SQL_LOGS=true`

### Supabase設定
- `SUPABASE_URL="https://xxxxxxxxxxxxxxxxxxxx.supabase.co"`
- `SUPABASE_ANON_KEY="eyxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`
- `SUPABASE_JWT_SECRET="CX/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`

