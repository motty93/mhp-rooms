.PHONY: build run dev test lint fmt clean

# バイナリ名
BINARY_NAME=mhp-rooms

# ビルドディレクトリ
BUILD_DIR=bin

# メインファイルのパス
MAIN_PATH=./cmd/server

# アプリケーションをビルド
build:
	@echo "ビルド中..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "ビルド完了: $(BUILD_DIR)/$(BINARY_NAME)"

# アプリケーションを実行
run: build
	@echo "アプリケーションを実行中..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 開発サーバーを起動（ホットリロードなし - 基本実装）
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

# 依存関係を取得
deps:
	@echo "依存関係を取得中..."
	@go mod tidy

# Docker開発環境コマンド
docker-up:
	@echo "Docker環境を起動中..."
	@docker-compose up -d

docker-down:
	@echo "Docker環境を停止中..."
	@docker-compose down

docker-build:
	@echo "Dockerイメージを再ビルド中..."
	@docker-compose build --no-cache app

docker-logs:
	@echo "アプリケーションログを表示中..."
	@docker-compose logs -f app

docker-reset:
	@echo "Docker環境をリセット中..."
	@docker-compose down -v
	@docker-compose up -d

# ヘルプを表示
help:
	@echo "利用可能なコマンド:"
	@echo "  build       - アプリケーションをビルド"
	@echo "  run         - アプリケーションを実行"
	@echo "  dev         - ホットリロード開発サーバーを起動"
	@echo "  test        - テストを実行"
	@echo "  lint        - リンターを実行"
	@echo "  fmt         - コードをフォーマット"
	@echo "  clean       - ビルド成果物をクリーンアップ"
	@echo "  deps        - 依存関係を取得"
	@echo "  docker-up   - Docker環境を起動"
	@echo "  docker-down - Docker環境を停止"
	@echo "  docker-build- Dockerイメージを再ビルド"
	@echo "  docker-logs - アプリケーションログを表示"
	@echo "  docker-reset- Docker環境をリセット"
	@echo "  help        - このヘルプを表示"
