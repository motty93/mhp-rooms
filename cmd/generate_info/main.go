package main

import (
	"fmt"
	"log"
	"os"

	"mhp-rooms/internal/info"
)

func main() {
	fmt.Println("=== 更新情報・ロードマップ 静的ファイル生成 ===")

	// コンテンツディレクトリの設定
	contentDirs := map[info.ArticleType]string{
		info.ArticleTypeRelease:     "content/info",
		info.ArticleTypeMaintenance: "content/info",
		info.ArticleTypeRoadmap:     "content/roadmap",
	}

	// 出力ディレクトリ
	outputDir := "static/generated/info"

	// ジェネレーター作成
	generator := info.NewGenerator(outputDir, contentDirs)

	// 生成実行
	if err := generator.Generate(); err != nil {
		log.Fatalf("エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ 全ての静的ファイルの生成が完了しました")
}
