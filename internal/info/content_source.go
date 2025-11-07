package info

// マークダウンコンテンツの取得元を表す
type ContentSource struct {
	Dir             string
	DefaultCategory ArticleType
}

// デフォルトのコンテンツソース設定を返す
func DefaultContentSources() []ContentSource {
	return []ContentSource{
		{Dir: "content/info"},
		{Dir: "content/roadmap", DefaultCategory: ArticleTypeRoadmap},
	}
}
