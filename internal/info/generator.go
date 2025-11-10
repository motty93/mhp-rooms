package info

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/feeds"
)

type Generator struct {
	parser         *Parser
	outputDir      string
	contentSources []ContentSource
}

func NewGenerator(outputDir string, sources []ContentSource) *Generator {
	cloned := make([]ContentSource, len(sources))
	copy(cloned, sources)
	return &Generator{
		parser:         NewParser(),
		outputDir:      outputDir,
		contentSources: cloned,
	}
}

func (g *Generator) Generate() error {
	// 出力ディレクトリ作成
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリ作成エラー: %w", err)
	}

	allArticles := make(ArticleList, 0)

	for _, source := range g.contentSources {
		articles, err := g.parser.ParseDirectory(source.Dir)
		if err != nil {
			return fmt.Errorf("%s ディレクトリのパースエラー: %w", source.Dir, err)
		}

		for _, article := range articles {
			if article.Category == "" && source.DefaultCategory != "" {
				article.Category = source.DefaultCategory
			}
		}

		allArticles = append(allArticles, articles...)
	}

	// 下書きを除外してソート
	publishedArticles := allArticles.ExcludeDrafts().SortByDateDesc()

	// JSONファイルとして保存（ハンドラーで使用）
	if err := g.generateJSON(publishedArticles); err != nil {
		return fmt.Errorf("JSON生成エラー: %w", err)
	}

	// RSSフィード生成
	if err := g.generateFeed(publishedArticles); err != nil {
		return fmt.Errorf("フィード生成エラー: %w", err)
	}

	fmt.Printf("✓ %d件の記事を処理しました\n", len(publishedArticles))
	return nil
}

// 記事データをJSONとして保存する
func (g *Generator) generateJSON(articles ArticleList) error {
	jsonPath := filepath.Join(g.outputDir, "articles.json")

	data, err := json.MarshalIndent(articles, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return err
	}

	fmt.Printf("✓ JSONファイルを生成しました: %s\n", jsonPath)

	return nil
}

// RSS/Atomフィードを生成する
func (g *Generator) generateFeed(articles ArticleList) error {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "HuntersHub 更新情報",
		Link:        &feeds.Link{Href: "https://huntershub.com"},
		Description: "モンスターハンターポータブルシリーズのアドホックパーティ ルーム管理サービス",
		Author:      &feeds.Author{Name: "HuntersHub Team"},
		Created:     now,
	}

	// 更新情報のみをフィードに含める（最新20件）
	infoArticles := articles.FilterByCategory(ArticleTypeNews)
	infoArticles = append(infoArticles, articles.FilterByCategory(ArticleTypeMaintenance)...)
	infoArticles = infoArticles.SortByDateDesc()

	maxItems := 20
	if len(infoArticles) < maxItems {
		maxItems = len(infoArticles)
	}

	for i := 0; i < maxItems; i++ {
		article := infoArticles[i]
		item := &feeds.Item{
			Title:       article.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("https://huntershub.com/info/%s", article.Slug)},
			Description: article.Summary,
			Content:     article.Content,
			Created:     article.Date,
		}

		if article.Updated != nil {
			item.Updated = *article.Updated
		} else {
			item.Updated = article.Date
		}

		feed.Items = append(feed.Items, item)
	}

	// RSS 2.0
	rssPath := filepath.Join(g.outputDir, "feed.xml")
	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	if err := os.WriteFile(rssPath, []byte(rss), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ RSSフィードを生成しました: %s\n", rssPath)

	// Atom
	atomPath := filepath.Join(g.outputDir, "atom.xml")
	atom, err := feed.ToAtom()
	if err != nil {
		return err
	}

	if err := os.WriteFile(atomPath, []byte(atom), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Atomフィードを生成しました: %s\n", atomPath)

	return nil
}
