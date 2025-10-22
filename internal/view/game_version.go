package view

// GetGameVersionColor はゲームバージョンに応じた色を返す
func GetGameVersionColor(code string) string {
	// 部屋一覧のCSSと同じ色設定に統一
	switch code {
	case "MHP":
		return "bg-amber-700" // 茶色系（rgba(139, 69, 19, 0.8)に近い）
	case "MHP2":
		return "bg-blue-600" // 青色系（rgba(70, 130, 180, 0.8)に近い）
	case "MHP2G":
		return "bg-green-700" // 緑色系（rgba(34, 139, 34, 0.8)に近い）
	case "MHP3":
		return "bg-yellow-600" // 金色系（rgba(218, 165, 32, 0.8)に近い）
	case "MHXX":
		return "bg-gray-900" // 黒色系（rgba(0, 0, 0, 0.8)に近い）
	default:
		return "bg-gray-600"
	}
}

// GetGameVersionIcon はゲームバージョンに応じたアイコンSVGを返す
func GetGameVersionIcon(code string) string {
	switch code {
	// Sony系
	case "MHP":
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">1</text>
		</svg>`
	case "MHP2":
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">2</text>
		</svg>`
	case "MHP2G":
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-sm font-bold">2G</text>
		</svg>`
	case "MHP3":
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">3</text>
		</svg>`
	// Nintendo系
	case "MHXX":
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xs font-bold">XX</text>
		</svg>`
	default:
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" />
		</svg>`
	}
}

// GetGameVersionAbbreviation はゲームバージョンの略称を返す
func GetGameVersionAbbreviation(code string) string {
	switch code {
	// Sony系
	case "MHP":
		return "1"
	case "MHP2":
		return "2"
	case "MHP2G":
		return "2G"
	case "MHP3":
		return "3"
	// Nintendo系
	case "MHXX":
		return "XX"
	default:
		return "?"
	}
}
