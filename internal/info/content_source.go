package info

// ContentSource はマークダウンコンテンツの取得元を表す
type ContentSource struct {
	Dir             string
	DefaultCategory ArticleType
}

// DefaultContentSources はデフォルトのコンテンツソース設定を返す
func DefaultContentSources() []ContentSource {
	return []ContentSource{
		{Dir: "content/info"},
		{Dir: "content/roadmap", DefaultCategory: ArticleTypeRoadmap},
	}
}
