package main

import (
	"fmt"
	"log"
	"os"

	"mhp-rooms/internal/info"
)

func main() {
	fmt.Println("=== 更新情報・ロードマップ・ブログ 静的ファイル生成 ===")

	// 更新情報・ロードマップ等の生成
	contentSources := info.DefaultContentSources()
	outputDir := "static/generated/info"
	generator := info.NewGenerator(outputDir, contentSources)

	if err := generator.Generate(); err != nil {
		log.Fatalf("エラー: %v\n", err)
		os.Exit(1)
	}

	// ブログ専用のJSON生成
	blogSources := []info.ContentSource{
		{Dir: "content/blog"},
	}
	blogOutputDir := "static/generated/blog"
	blogGenerator := info.NewGenerator(blogOutputDir, blogSources)

	if err := blogGenerator.Generate(); err != nil {
		log.Fatalf("ブログ生成エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ 全ての静的ファイルの生成が完了しました")
}
