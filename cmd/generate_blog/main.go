package main

import (
	"fmt"
	"log"
	"os"

	"mhp-rooms/internal/info"
)

func main() {
	fmt.Println("=== ブログ記事 静的ファイル生成 ===")

	blogSources := []info.ContentSource{
		{Dir: "content/blog", DefaultCategory: info.ArticleTypeBlogTechnical},
	}

	outputDir := "static/generated/blog"

	generator := info.NewGenerator(outputDir, blogSources)

	if err := generator.Generate(); err != nil {
		log.Fatalf("エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ ブログ記事の静的ファイル生成が完了しました")
}
