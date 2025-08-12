package handlers

import (
	"log"
	"net/http"

	"mhp-rooms/internal/repository"
)

type ProfileHandler struct {
	BaseHandler
	logger *log.Logger
}

func NewProfileHandler(repo *repository.Repository) *ProfileHandler {
	return &ProfileHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		logger: log.New(log.Writer(), "[ProfileHandler] ", log.LstdFlags),
	}
}

// Profile プロフィールページを表示
func (ph *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title: "プロフィール",
	}

	renderTemplate(w, "profile.tmpl", data)
}

// EditForm プロフィール編集フォームを返す（htmx用）
func (ph *ProfileHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	// モック用のダミーレスポンス
	html := `
	<div class="p-6">
		<h3 class="text-xl font-bold text-gray-800 mb-4">プロフィール編集</h3>
		<div class="space-y-4">
			<div>
				<label class="text-sm font-bold text-gray-600 block mb-1">ユーザー名</label>
				<input type="text" value="rdwbocungelt5" class="w-full bg-gray-100 text-gray-800 rounded-lg p-2 border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
			</div>
			<div>
				<label class="text-sm font-bold text-gray-600 block mb-1">自己紹介</label>
				<textarea class="w-full bg-gray-100 text-gray-800 rounded-lg p-2 border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" rows="4">モンハン大好きです！一緒に楽しく狩りに行けるフレンドを募集しています。VCも可能です。よろしくお願いします！</textarea>
			</div>
			<div class="flex space-x-2">
				<button class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">保存</button>
				<button hx-get="/api/profile/view" hx-target="#profile-card" hx-swap="outerHTML" class="w-full bg-gray-200 hover:bg-gray-300 text-gray-800 font-bold py-2 px-4 rounded-lg transition-colors">キャンセル</button>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Activity アクティビティタブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Activity(w http.ResponseWriter, r *http.Request) {
	// モック用のダミーレスポンス
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">最近の活動履歴</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg flex justify-between items-center hover:bg-gray-100 transition-colors border border-gray-200">
				<div class="flex items-center space-x-4">
					<i class="fa-solid fa-door-open text-green-500"></i>
					<div>
						<p class="font-semibold text-gray-800">【部屋作成】古龍種連戦</p>
						<p class="text-sm text-gray-500">ターゲット: クシャルダオラ</p>
					</div>
				</div>
				<span class="text-xs text-gray-400">3時間前</span>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg flex justify-between items-center hover:bg-gray-100 transition-colors border border-gray-200">
				<div class="flex items-center space-x-4">
					<i class="fa-solid fa-right-to-bracket text-blue-500"></i>
					<div>
						<p class="font-semibold text-gray-800">【部屋参加】二つ名持ちモンスター</p>
						<p class="text-sm text-gray-500">ホスト: 素材コレクター</p>
					</div>
				</div>
				<span class="text-xs text-gray-400">昨日</span>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Rooms 作成した部屋タブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Rooms(w http.ResponseWriter, r *http.Request) {
	// モック用のダミーレスポンス
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">作成した部屋</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg border border-gray-200">
				<div class="flex justify-between items-start mb-2">
					<h4 class="font-semibold text-gray-800">テスト部屋（更新済み）</h4>
					<span class="text-xs text-gray-500">3時間前</span>
				</div>
				<p class="text-sm text-gray-600 mb-2">部屋設定の更新機能が正常に動作することを確認しました。</p>
				<div class="flex items-center space-x-4 text-xs text-gray-500">
					<span>MHP3</span>
					<span>1/4人</span>
					<span class="text-green-600">アクティブ</span>
				</div>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg border border-gray-200">
				<div class="flex justify-between items-start mb-2">
					<h4 class="font-semibold text-gray-800">古龍種連戦</h4>
					<span class="text-xs text-gray-500">1日前</span>
				</div>
				<p class="text-sm text-gray-600 mb-2">古龍種を順番に討伐していきます</p>
				<div class="flex items-center space-x-4 text-xs text-gray-500">
					<span>MHP2G</span>
					<span>0/4人</span>
					<span class="text-red-600">終了</span>
				</div>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Friends フレンドタブコンテンツを返す（htmx用）
func (ph *ProfileHandler) Friends(w http.ResponseWriter, r *http.Request) {
	// モック用のダミーレスポンス
	html := `
	<div>
		<h3 class="text-xl font-bold mb-4 text-gray-800">フレンドリスト</h3>
		<div class="space-y-4">
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="https://placehold.co/40x40/3b82f6/ffffff?text=H" alt="ハンター太郎">
					<div>
						<p class="font-semibold text-gray-800">ハンター太郎</p>
						<p class="text-xs text-gray-500">オンライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-green-500 rounded-full"></span>
					<span class="text-xs text-gray-500">2日前にフレンドになりました</span>
				</div>
			</div>
			<div class="bg-gray-50 p-4 rounded-lg flex items-center justify-between border border-gray-200">
				<div class="flex items-center space-x-4">
					<img class="w-10 h-10 rounded-full object-cover" src="https://placehold.co/40x40/ef4444/ffffff?text=S" alt="素材コレクター">
					<div>
						<p class="font-semibold text-gray-800">素材コレクター</p>
						<p class="text-xs text-gray-500">オフライン</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="w-3 h-3 bg-gray-400 rounded-full"></span>
					<span class="text-xs text-gray-500">5日前にフレンドになりました</span>
				</div>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}