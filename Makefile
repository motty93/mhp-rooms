.PHONY: build run dev test lint fmt clean migrate migrate-dev container-up container-down setup generate-ogp generate-info

# バイナリ名
BINARY_NAME=mhp-rooms
MIGRATE_BINARY=migrate

# ビルドディレクトリ
BUILD_DIR=bin

# メインファイルのパス
MAIN_PATH=./cmd/server
MIGRATE_PATH=./cmd/migrate
SEED_PATH=./cmd/seed
GENERATE_INFO_PATH=./cmd/generate_info

# アプリケーションをビルド
build: generate-info
	@echo "ビルド中..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@go build -o $(BUILD_DIR)/$(MIGRATE_BINARY) $(MIGRATE_PATH)
	@echo "ビルド完了: $(BUILD_DIR)/$(BINARY_NAME), $(BUILD_DIR)/$(MIGRATE_BINARY)"

# アプリケーションを実行
run: build
	@echo "アプリケーションを実行中..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 開発サーバーを起動（air使用）
dev:
	@echo "ホットリロード開発サーバーを起動中..."
	@air

# テストを実行
test:
	@echo "テストを実行中..."
	@go test -v ./...

# リンターを実行
lint:
	@echo "リンターを実行中..."
	@go vet ./...
	@gofmt -s -l .

# コードをフォーマット
fmt:
	@echo "コードをフォーマット中..."
	@go fmt ./...
	@prettier --write templates/**/*.tmpl

# ビルド成果物をクリーンアップ
clean:
	@echo "クリーンアップ中..."
	@rm -rf $(BUILD_DIR)
	@echo "クリーンアップ完了"

# マイグレーション
migrate: build
	@echo "マイグレーションを実行中..."
	@./$(BUILD_DIR)/$(MIGRATE_BINARY) -migrate

migrate-dev:
	@echo "開発モードでマイグレーションを実行中..."
	@go run $(MIGRATE_PATH)/main.go -migrate

# seeds
seeds:
	@echo "シードデータを挿入中..."
	@go run $(SEED_PATH)/main.go -seed

# アクティビティデータ修正（一時的なデータマイグレーション）
fix-activity:
	@echo "アクティビティデータを修正中..."
	@echo "⚠️  このスクリプトは既に修正済みのデータをスキップします"
	@go run ./cmd/fix-activity/main.go
	@echo "✅ アクティビティデータ修正完了"

# OGP画像生成
generate-ogp:
	@if [ -z "$(ROOM_ID)" ]; then \
		echo "エラー: ROOM_ID を指定してください"; \
		echo "使用例: make generate-ogp ROOM_ID=<uuid>"; \
		exit 1; \
	fi
	@echo "OGP画像を生成中: ROOM_ID=$(ROOM_ID)"
	@ROOM_ID=$(ROOM_ID) \
	OG_PREFIX=dev \
	go run cmd/ogp-renderer/main.go
	@echo "✅ OGP画像生成完了: tmp/images/dev/ogp/rooms/$(ROOM_ID).png"

# 更新情報・ロードマップの静的ファイル生成
generate-info:
	@echo "更新情報・ロードマップの静的ファイルを生成中..."
	@go run $(GENERATE_INFO_PATH)/main.go
	@echo "✅ 静的ファイル生成完了"

# 依存関係を取得
deps:
	@echo "依存関係を取得中..."
	@go mod tidy

# 初期設定（開発環境セットアップ）
setup: deps container-up migrate-dev seeds
	@echo ""
	@echo "✅ 初期設定が完了しました！"
	@echo ""
	@echo "次のコマンドで開発サーバーを起動できます:"
	@echo "  make dev"
	@echo ""
	@echo "アクセスURL: http://localhost:8080"

# Docker開発環境コマンド（app/dbコンテナ）
container-up:
	@echo "Dockerコンテナ(app/db)を起動中..."
	@docker compose -f compose.db.yml up -d

container-down:
	@echo "コンテナを停止中..."
	@docker compose down

container-logs:
	@echo "コンテナログを表示中..."
	@docker compose logs -f

container-reset:
	@echo "コンテナ環境をリセット中..."
	@docker compose down -v
	@docker compose up -d

# 旧コマンド（互換性のため残す）
docker-up: container-up
docker-down: container-down

# ヘルプを表示
help:
	@echo "利用可能なコマンド:"
	@echo "  setup         - 🚀 初期設定（開発環境の完全セットアップ）"
	@echo "  build         - アプリケーションをビルド（更新情報も自動生成）"
	@echo "  run           - アプリケーションを実行"
	@echo "  dev           - ホットリロード開発サーバーを起動（air使用）"
	@echo "  migrate       - マイグレーションを実行"
	@echo "  migrate-dev   - 開発モードでマイグレーションを実行"
	@echo "  seeds         - シードデータを挿入"
	@echo "  fix-activity  - アクティビティデータを修正（一時的なデータマイグレーション）"
	@echo "  generate-ogp  - OGP画像を生成（ROOM_ID=<uuid>を指定）"
	@echo "  generate-info - 更新情報・ロードマップの静的ファイルを生成"
	@echo "  test          - テストを実行"
	@echo "  lint          - リンターを実行"
	@echo "  fmt           - コードをフォーマット"
	@echo "  clean         - ビルド成果物をクリーンアップ"
	@echo "  deps          - 依存関係を取得"
	@echo "  container-up  - DBコンテナを起動"
	@echo "  container-down- コンテナを停止"
	@echo "  container-logs- コンテナログを表示"
	@echo "  container-reset- コンテナ環境をリセット"
	@echo "  help          - このヘルプを表示"
