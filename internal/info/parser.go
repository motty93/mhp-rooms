package info

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v2"
)

// Parser はマークダウンファイルをパースする
type Parser struct {
	md goldmark.Markdown
}

// NewParser は新しいParserを作成する
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	return &Parser{md: md}
}

// ParseFile はマークダウンファイルをパースしてArticleを返す
func (p *Parser) ParseFile(filePath string) (*Article, error) {
	// ファイル読み込み
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	// パーサーコンテキスト作成
	ctx := parser.NewContext()
	var buf bytes.Buffer

	// マークダウンをHTMLに変換
	if err := p.md.Convert(content, &buf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("マークダウン変換エラー: %w", err)
	}

	// メタデータ取得
	metaData := meta.Get(ctx)
	if metaData == nil {
		return nil, fmt.Errorf("frontmatterが見つかりません: %s", filePath)
	}

	// Article構造体にマッピング
	article := &Article{
		Content:  buf.String(),
		FilePath: filePath,
	}

	// YAMLデータをArticle構造体にマッピング
	yamlData, err := yaml.Marshal(metaData)
	if err != nil {
		return nil, fmt.Errorf("メタデータのマーシャルエラー: %w", err)
	}

	if err := yaml.Unmarshal(yamlData, article); err != nil {
		return nil, fmt.Errorf("メタデータのアンマーシャルエラー: %w", err)
	}

	return article, nil
}

// ParseDirectory はディレクトリ内の全マークダウンファイルをパースする
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
				// エラーはログに出すが処理は継続
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
