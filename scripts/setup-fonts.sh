#!/bin/bash

# OGP画像生成用フォントのダウンロードスクリプト

set -e

FONT_DIR="cmd/ogp-renderer/assets/fonts"
BOLD_URL="https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Bold.ttf"
REGULAR_URL="https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP-Regular.ttf"

echo "フォントファイルをダウンロードします..."

# ディレクトリ作成
mkdir -p "$FONT_DIR"

# NotoSansJP Bold
if [ -f "$FONT_DIR/NotoSansJP-Bold.ttf" ]; then
    echo "✓ NotoSansJP-Bold.ttf は既に存在します"
else
    echo "  NotoSansJP-Bold.ttf をダウンロード中..."
    curl -L "$BOLD_URL" -o "$FONT_DIR/NotoSansJP-Bold.ttf"
    echo "✓ NotoSansJP-Bold.ttf をダウンロードしました"
fi

# NotoSansJP Regular
if [ -f "$FONT_DIR/NotoSansJP-Regular.ttf" ]; then
    echo "✓ NotoSansJP-Regular.ttf は既に存在します"
else
    echo "  NotoSansJP-Regular.ttf をダウンロード中..."
    curl -L "$REGULAR_URL" -o "$FONT_DIR/NotoSansJP-Regular.ttf"
    echo "✓ NotoSansJP-Regular.ttf をダウンロードしました"
fi

echo ""
echo "フォントのダウンロードが完了しました。"
echo "これらのフォントファイルをコミットしてください:"
echo "  git add $FONT_DIR/*.ttf"
echo "  git commit -m \"feat: Add NotoSansJP fonts for OGP image generation\""
