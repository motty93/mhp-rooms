package info

import "time"

// ArticleType は記事の種類を表す
type ArticleType string

const (
	ArticleTypeRelease     ArticleType = "release"
	ArticleTypeMaintenance ArticleType = "maintenance"
	ArticleTypeFeature     ArticleType = "feature"
	ArticleTypeRoadmap     ArticleType = "roadmap"
)

// Article は更新情報やロードマップの記事を表す
type Article struct {
	Title    string      `yaml:"title"`
	Slug     string      `yaml:"slug"`
	Date     time.Time   `yaml:"date"`
	Updated  *time.Time  `yaml:"updated"`
	Category ArticleType `yaml:"category"`
	Summary  string      `yaml:"summary"`
	Tags     []string    `yaml:"tags"`
	Draft    bool        `yaml:"draft"`
	Status   string      `yaml:"status"` // ロードマップ用（planned, in_progress, completed）
	Content  string      // マークダウンから変換されたHTML
	FilePath string      // 元のマークダウンファイルパス
}

// ArticleList は記事のリスト
type ArticleList []*Article

// FilterByCategory はカテゴリーでフィルタリングする
func (al ArticleList) FilterByCategory(category ArticleType) ArticleList {
	var filtered ArticleList
	for _, article := range al {
		if article.Category == category {
			filtered = append(filtered, article)
		}
	}
	return filtered
}

// ExcludeDrafts は下書きを除外する
func (al ArticleList) ExcludeDrafts() ArticleList {
	var published ArticleList
	for _, article := range al {
		if !article.Draft {
			published = append(published, article)
		}
	}
	return published
}

// SortByDateDesc は日付の降順でソートする
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
