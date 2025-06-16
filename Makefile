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
	@echo "開発サーバーを起動中..."
	@go run $(MAIN_PATH)

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

# ヘルプを表示
help:
	@echo "利用可能なコマンド:"
	@echo "  build    - アプリケーションをビルド"
	@echo "  run      - アプリケーションを実行"
	@echo "  dev      - 開発サーバーを起動"
	@echo "  test     - テストを実行"
	@echo "  lint     - リンターを実行"
	@echo "  fmt      - コードをフォーマット"
	@echo "  clean    - ビルド成果物をクリーンアップ"
	@echo "  deps     - 依存関係を取得"
	@echo "  help     - このヘルプを表示"