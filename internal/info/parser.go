package info

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v2"
)

// マークダウンファイルをパースする
type Parser struct {
	md goldmark.Markdown
}

func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	return &Parser{md: md}
}

// マークダウンファイルをパースしてArticleを返す
func (p *Parser) ParseFile(filePath string) (*Article, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	ctx := parser.NewContext()
	var buf bytes.Buffer

	if err := p.md.Convert(content, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("マークダウン変換エラー: %w", err)
	}

	metaData := meta.Get(ctx)
	if metaData == nil {
		return nil, fmt.Errorf("frontmatterが見つかりません: %s", filePath)
	}

	article := &Article{
		Content:  buf.String(),
		FilePath: filePath,
	}

	yamlData, err := yaml.Marshal(metaData)
	if err != nil {
		return nil, fmt.Errorf("メタデータのマーシャルエラー: %w", err)
	}

	if err := yaml.Unmarshal(yamlData, article); err != nil {
		return nil, fmt.Errorf("メタデータのアンマーシャルエラー: %w", err)
	}

	// 読了時間を計算
	article.ReadingTime = calculateReadingTime(string(content))

	return article, nil
}

// calculateReadingTime は記事の読了時間を計算する（日本語: 500文字/分）
func calculateReadingTime(content string) int {
	// マークダウンのメタデータ部分（---で囲まれた部分）を除外
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		content = parts[2]
	}

	// マークダウン記法を除外（簡易版）
	content = strings.ReplaceAll(content, "#", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "_", "")
	content = strings.ReplaceAll(content, "`", "")
	content = strings.ReplaceAll(content, "[", "")
	content = strings.ReplaceAll(content, "]", "")
	content = strings.ReplaceAll(content, "(", "")
	content = strings.ReplaceAll(content, ")", "")

	// 文字数をカウント
	charCount := utf8.RuneCountInString(content)

	// 日本語の平均読書速度: 500文字/分
	minutes := charCount / 500
	if minutes < 1 {
		return 1
	}

	return minutes
}

// ディレクトリ内の全マークダウンファイルをパースする
func (p *Parser) ParseDirectory(dirPath string) (ArticleList, error) {
	var articles ArticleList

	// ディレクトリが存在しない場合は空のリストを返す
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return articles, nil
	}

	// ディレクトリ内のファイルを走査
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .mdファイルのみ処理
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			article, parseErr := p.ParseFile(path)

			if parseErr != nil {
				fmt.Printf("警告: %sのパースに失敗: %v\n", path, parseErr)
				return nil
			}

			articles = append(articles, article)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ディレクトリ走査エラー: %w", err)
	}

	return articles, nil
}
