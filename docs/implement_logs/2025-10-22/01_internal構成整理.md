# internal配下の構造整理

**実装日**: 2025年10月22日  
**Issue**: なし  
**実装時間**: 約3時間

## 概要

service層導入前提で散らばっていたテンプレート補助やインフラ系コードを整理し、役割ごとの境界を明確化した。影響範囲は Go コードのみで、外部 API や DB スキーマの変更はない。

## 実装内容

### 1. ビューレイヤーの集約
- **ファイル**: `internal/view/funcs.go`, `internal/view/render.go`, `internal/view/palette.go`, `internal/view/game_version.go`, `internal/handlers/handlers.go`, `internal/handlers/room_detail.go`
- **内容**:
  - 旧 `internal/helpers`, `internal/render`, `internal/palette` を統合し、新たに `view` パッケージを作成。
  - テンプレート関数・レンダリング処理・ゲームバージョンの配色/表示ロジックを単一モジュールにまとめ、各ハンドラーは `view.Template` / `view.Partial` を利用するよう変更。
- **目的**: プレゼンテーション層の責務を明確化し、テンプレート変更時の依存追跡を簡素化。

### 2. インフラ層の配置変更
- **ファイル**: `internal/infrastructure/storage/gcs.go`, `internal/infrastructure/sse/hub.go`, `cmd/server/application.go`, `internal/handlers/profile.go`, `internal/handlers/report.go`, `internal/handlers/room_messages.go`, `internal/handlers/rooms.go`
- **内容**:
  - GCS アップローダーと SSE Hub を `internal/infrastructure` 配下へ移動し、呼び出し側のインポートを更新。
  - 既存の `internal/storage` と `internal/sse` ディレクトリを廃止。
- **目的**: 永続化・外部接続といったインフラ依存コードを一箇所に集約し、層の依存方向を明示。

### 3. Discord 連携のモジュール化
- **ファイル**: `internal/integration/discord/webhook.go`, `internal/handlers/contact.go`
- **内容**:
  - Discord Webhook 通知ロジックを `integration/discord` パッケージとして切り出し。
  - お問い合わせハンドラーは新パッケージを利用。
- **目的**: 外部サービス連携を `utils` から分離し、アプリ固有ヘルパーと共通関数を明確に区別。

### 4. utils のスリム化
- **ファイル**: `internal/utils/game_version_helper.go`（削除）
- **内容**:
  - テンプレート専用のゲームバージョン表示処理を `view` に移管し、`utils` から削除。
- **目的**: 汎用的なユーティリティと UI 依存コードの混在を解消。

## テスト

- `GOCACHE=$(pwd)/.gocache go test ./...`

## メモ

- README 内のパッケージ構成図は未更新。開発チームで共有しているアーキテクチャ資料を後続で修正予定。
