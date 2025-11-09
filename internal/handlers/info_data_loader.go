package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"mhp-rooms/internal/info"
)

// loadArticlesWithFallback はJSONから記事を読み込み、ファイルが無ければジェネレーターで再生成する
func loadArticlesWithFallback(path string, generator *info.Generator) (info.ArticleList, error) {
	articles, err := readArticlesFromJSON(path)
	if err == nil {
		return articles, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		if generator == nil {
			return nil, fmt.Errorf("記事データが存在しません: %w", err)
		}
		if genErr := generator.Generate(); genErr != nil {
			return nil, fmt.Errorf("記事データの生成に失敗しました: %w", genErr)
		}
		return readArticlesFromJSON(path)
	}

	return nil, err
}

func readArticlesFromJSON(path string) (info.ArticleList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var articles info.ArticleList
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, err
	}

	return articles, nil
}

func readGeneratedFile(path string, generator *info.Generator) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		if generator == nil {
			return nil, err
		}
		if genErr := generator.Generate(); genErr != nil {
			return nil, fmt.Errorf("生成ファイルの作成に失敗しました: %w", genErr)
		}
		return os.ReadFile(path)
	}

	return nil, err
}
