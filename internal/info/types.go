package info

import "time"

type ArticleType string

const (
	ArticleTypeNews        ArticleType = "news"
	ArticleTypeRelease     ArticleType = "release"
	ArticleTypeMaintenance ArticleType = "maintenance"
	ArticleTypeFeature     ArticleType = "feature"
	ArticleTypeRoadmap     ArticleType = "roadmap"
	ArticleTypeOperator    ArticleType = "operator"
	ArticleTypeGuide       ArticleType = "guide"
	ArticleTypeFAQ         ArticleType = "faq"
	ArticleTypeTerms       ArticleType = "terms"
	ArticleTypePrivacy     ArticleType = "privacy"
	// ブログ用カテゴリ
	ArticleTypeBlogCommunity       ArticleType = "community"
	ArticleTypeBlogTechnical       ArticleType = "technical"
	ArticleTypeBlogTroubleshooting ArticleType = "troubleshooting"
)

type Article struct {
	Title       string      `yaml:"title"`
	Slug        string      `yaml:"slug"`
	Date        time.Time   `yaml:"date"`
	Updated     *time.Time  `yaml:"updated"`
	Category    ArticleType `yaml:"category"`
	Summary     string      `yaml:"summary"`
	Draft       bool        `yaml:"draft"`
	Status      string      `yaml:"status"` // ロードマップ用（planned, in_progress, completed）
	Images      []string    `yaml:"images"` // 記事内で使用する画像パスのリスト
	Content     string      // マークダウンから変換されたHTML
	FilePath    string      // 元のマークダウンファイルパス
	ReadingTime int         // 読了時間（分）
}

type ArticleList []*Article

func (al ArticleList) FilterByCategory(category ArticleType) ArticleList {
	var filtered ArticleList

	for _, article := range al {
		if article.Category == category {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

func (al ArticleList) ExcludeDrafts() ArticleList {
	var published ArticleList

	for _, article := range al {
		if !article.Draft {
			published = append(published, article)
		}
	}

	return published
}

func (al ArticleList) SortByDateDesc() ArticleList {
	// シンプルなバブルソート（記事数が少ないため）
	sorted := make(ArticleList, len(al))
	copy(sorted, al)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Date.Before(sorted[j].Date) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
