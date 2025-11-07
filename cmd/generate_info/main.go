package main

import (
	"fmt"
	"log"
	"os"

	"mhp-rooms/internal/info"
)

func main() {
	fmt.Println("=== 更新情報・ロードマップ 静的ファイル生成 ===")

	contentSources := info.DefaultContentSources()

	outputDir := "static/generated/info"

	generator := info.NewGenerator(outputDir, contentSources)

	if err := generator.Generate(); err != nil {
		log.Fatalf("エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ 全ての静的ファイルの生成が完了しました")
}
