.PHONY: build run dev test lint fmt clean migrate migrate-dev container-up container-down

# バイナリ名
BINARY_NAME=mhp-rooms
MIGRATE_BINARY=migrate

# ビルドディレクトリ
BUILD_DIR=bin

# メインファイルのパス
MAIN_PATH=./cmd/server
MIGRATE_PATH=./cmd/migrate

# アプリケーションをビルド
build:
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

# 依存関係を取得
deps:
	@echo "依存関係を取得中..."
	@go mod tidy

# Docker開発環境コマンド（app以外のコンテナのみ）
container-up:
	@echo "DBとRedisコンテナを起動中..."
	@docker compose up -d

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
	@echo "  build         - アプリケーションをビルド"
	@echo "  run           - アプリケーションを実行"
	@echo "  dev           - ホットリロード開発サーバーを起動（air使用）"
	@echo "  migrate       - マイグレーションを実行"
	@echo "  migrate-dev   - 開発モードでマイグレーションを実行"
	@echo "  test          - テストを実行"
	@echo "  lint          - リンターを実行"
	@echo "  fmt           - コードをフォーマット"
	@echo "  clean         - ビルド成果物をクリーンアップ"
	@echo "  deps          - 依存関係を取得"
	@echo "  container-up  - DBとRedisコンテナを起動"
	@echo "  container-down- コンテナを停止"
	@echo "  container-logs- コンテナログを表示"
	@echo "  container-reset- コンテナ環境をリセット"
	@echo "  help          - このヘルプを表示"
