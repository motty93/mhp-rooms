package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

// TemplateData はテンプレートに渡すデータ構造体
type TemplateData struct {
	Title    string
	HasHero  bool
	User     interface{} // 将来的にユーザー情報を格納
	PageData interface{} // ページ固有のデータ
}

// renderTemplate はテンプレートをレンダリングする共通関数
func renderTemplate(w http.ResponseWriter, templateName string, data TemplateData) {
	// テンプレート関数を定義
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	// 必要なテンプレートファイルを全て読み込み
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "layouts", "base.html"),
		filepath.Join("templates", "components", "header.html"),
		filepath.Join("templates", "components", "footer.html"),
		filepath.Join("templates", "pages", templateName),
	)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded block">
    <strong>Hello World!</strong> Go + HTMX + Tailwind CSS + Air ホットリロードで動作しています！
</div>
<script>
    document.getElementById('hello-result').classList.remove('hidden');
</script>`)
}
